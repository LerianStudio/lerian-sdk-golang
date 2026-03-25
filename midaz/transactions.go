// transactions.go implements the transactionsServiceAPI for managing atomic
// financial movements within a ledger. Transactions are composed of one or
// more operations (debit/credit legs) and follow a state machine: once
// created they can be committed, cancelled, or reverted but never deleted.
package midaz

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/url"
	"strings"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/core"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/LerianStudio/lerian-sdk-golang/pkg/pagination"
)

// transactionsServiceAPI provides CRUD operations (minus Delete) and state
// machine actions for transactions scoped to an organization and ledger.
// Transactions are immutable once created and cannot be deleted; instead
// they transition through states via Commit, Cancel, and Revert actions.
type transactionsServiceAPI interface {
	// Create creates a new transaction within the specified organization and ledger.
	Create(ctx context.Context, orgID, ledgerID string, input *CreateTransactionInput) (*Transaction, error)

	// Get retrieves a transaction by its unique identifier.
	Get(ctx context.Context, orgID, ledgerID, id string) (*Transaction, error)

	// List returns a paginated iterator over transactions in a ledger.
	List(ctx context.Context, orgID, ledgerID string, opts *models.CursorListOptions) *pagination.Iterator[Transaction]

	// Update partially updates an existing transaction (e.g., description or metadata).
	Update(ctx context.Context, orgID, ledgerID, id string, input *UpdateTransactionInput) (*Transaction, error)

	// Commit transitions the transaction to committed state, finalizing
	// all its operations and applying balance changes.
	Commit(ctx context.Context, orgID, ledgerID, id string) (*Transaction, error)

	// Cancel transitions the transaction to cancelled state, releasing
	// any held balances without applying the operations.
	Cancel(ctx context.Context, orgID, ledgerID, id string) (*Transaction, error)

	// Revert creates a reversal of a previously committed transaction,
	// undoing its balance effects with compensating operations.
	Revert(ctx context.Context, orgID, ledgerID, id string) (*Transaction, error)

	// CreateAnnotation creates a new transaction using annotation format.
	CreateAnnotation(ctx context.Context, orgID, ledgerID string, input *CreateTransactionInput) (*Transaction, error)

	// CreateDSL creates a new transaction using a DSL file.
	CreateDSL(ctx context.Context, orgID, ledgerID string, dslContent []byte) (*Transaction, error)

	// CreateInflow creates a new inflow transaction.
	CreateInflow(ctx context.Context, orgID, ledgerID string, input *CreateTransactionInflowInput) (*Transaction, error)

	// CreateOutflow creates a new outflow transaction.
	CreateOutflow(ctx context.Context, orgID, ledgerID string, input *CreateTransactionOutflowInput) (*Transaction, error)
}

// transactionsService is the concrete implementation of [transactionsServiceAPI].
// It embeds [core.BaseService] to inherit the HTTP transport layer.
type transactionsService struct {
	core.BaseService
}

// newTransactionsService creates a new [transactionsServiceAPI] backed by the
// given transaction [core.Backend].
func newTransactionsService(backend core.Backend) transactionsServiceAPI {
	return &transactionsService{
		BaseService: core.BaseService{Backend: backend},
	}
}

// Compile-time interface compliance check.
var _ transactionsServiceAPI = (*transactionsService)(nil)

// transactionsBasePath builds the transactions collection path for the given
// organization and ledger.
func transactionsBasePath(orgID, ledgerID string) string {
	return "/organizations/" + url.PathEscape(orgID) + "/ledgers/" + url.PathEscape(ledgerID) + "/transactions"
}

// transactionsItemPath builds the path for a specific transaction.
func transactionsItemPath(orgID, ledgerID, id string) string {
	return transactionsBasePath(orgID, ledgerID) + "/" + url.PathEscape(id)
}

const (
	transactionResource = "Transaction"
	maxDSLContentSize   = 10 << 20
)

type createTransactionRequest struct {
	Description              *string          `json:"description,omitempty"`
	Code                     *string          `json:"code,omitempty"`
	ChartOfAccountsGroupName *string          `json:"chartOfAccountsGroupName,omitempty"`
	ParentTransactionID      *string          `json:"parentTransactionId,omitempty"`
	Pending                  *bool            `json:"pending,omitempty"`
	Route                    *string          `json:"route,omitempty"`
	TransactionDate          *string          `json:"transactionDate,omitempty"`
	Send                     *TransactionSend `json:"send,omitempty"`
	Metadata                 models.Metadata  `json:"metadata,omitempty"`
}

func createDSLMultipartBody(dslContent []byte) ([]byte, string, error) {
	var buffer bytes.Buffer

	writer := multipart.NewWriter(&buffer)

	part, err := writer.CreateFormFile("transaction", "transaction.dsl")
	if err != nil {
		return nil, "", err
	}

	if _, err := part.Write(dslContent); err != nil {
		return nil, "", err
	}

	if err := writer.Close(); err != nil {
		return nil, "", err
	}

	return buffer.Bytes(), writer.FormDataContentType(), nil
}

func buildCreateTransactionRequest(operation string, input *CreateTransactionInput) (*createTransactionRequest, error) {
	if err := validateTransactionSend(operation, input.Send); err != nil {
		return nil, err
	}

	request := &createTransactionRequest{
		Description:              input.Description,
		Code:                     input.Code,
		ChartOfAccountsGroupName: input.ChartOfAccountsGroupName,
		ParentTransactionID:      input.ParentTransactionID,
		Pending:                  input.Pending,
		Route:                    input.Route,
		TransactionDate:          input.TransactionDate,
		Send:                     input.Send,
		Metadata:                 input.Metadata,
	}

	return request, nil
}

func validateTransactionSend(operation string, send *TransactionSend) error {
	if send == nil {
		return sdkerrors.NewValidation(operation, transactionResource, "send is required")
	}

	if strings.TrimSpace(send.Asset) == "" {
		return sdkerrors.NewValidation(operation, transactionResource, "send asset is required")
	}

	if strings.TrimSpace(send.Value) == "" {
		return sdkerrors.NewValidation(operation, transactionResource, "send value is required")
	}

	if err := validateTransactionVariantLegs(operation, "source", send.Source.From); err != nil {
		return err
	}

	return validateTransactionVariantLegs(operation, "distribute", send.Distribute.To)
}

func validateTransactionVariantLegs(operation, field string, legs []TransactionOperationLeg) error {
	if len(legs) == 0 {
		return sdkerrors.NewValidation(operation, transactionResource, field+" legs are required")
	}

	for _, leg := range legs {
		if strings.TrimSpace(leg.AccountAlias) == "" && strings.TrimSpace(leg.BalanceKey) == "" {
			return sdkerrors.NewValidation(operation, transactionResource, field+" leg identifier is required")
		}
	}

	return nil
}

// Create creates a new transaction within the specified organization and ledger.
func (s *transactionsService) Create(ctx context.Context, orgID, ledgerID string, input *CreateTransactionInput) (*Transaction, error) {
	const operation = "Transactions.Create"

	if err := ensureService(s); err != nil {
		return nil, err
	}

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "ledger id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "input is required")
	}

	request, err := buildCreateTransactionRequest(operation, input)
	if err != nil {
		return nil, err
	}

	return core.Create[Transaction, createTransactionRequest](ctx, &s.BaseService, transactionsBasePath(orgID, ledgerID), request)
}

// CreateAnnotation creates a new transaction using annotation format.
func (s *transactionsService) CreateAnnotation(ctx context.Context, orgID, ledgerID string, input *CreateTransactionInput) (*Transaction, error) {
	const operation = "Transactions.CreateAnnotation"

	if err := ensureService(s); err != nil {
		return nil, err
	}

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "ledger id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "input is required")
	}

	request, err := buildCreateTransactionRequest(operation, input)
	if err != nil {
		return nil, err
	}

	return core.Create[Transaction, createTransactionRequest](ctx, &s.BaseService, transactionsBasePath(orgID, ledgerID)+"/annotation", request)
}

// CreateDSL creates a new transaction using a DSL file.
func (s *transactionsService) CreateDSL(ctx context.Context, orgID, ledgerID string, dslContent []byte) (*Transaction, error) {
	const operation = "Transactions.CreateDSL"

	if err := ensureService(s); err != nil {
		return nil, err
	}

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "ledger id is required")
	}

	if len(dslContent) == 0 {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "DSL content is required")
	}

	if len(dslContent) > maxDSLContentSize {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "DSL content exceeds maximum allowed size")
	}

	body, contentType, err := createDSLMultipartBody(dslContent)
	if err != nil {
		return nil, sdkerrors.NewInternal("midaz", operation, "failed to build DSL multipart request", err)
	}

	backend, err := core.ResolveBackend(&s.BaseService)
	if err != nil {
		return nil, err
	}

	res, err := backend.Do(ctx, core.Request{
		Method:      "POST",
		Path:        transactionsBasePath(orgID, ledgerID) + "/dsl",
		BodyBytes:   body,
		ContentType: contentType,
	})
	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, sdkerrors.NewInternal("midaz", operation, "backend returned nil response", nil)
	}

	var result Transaction
	if err := json.Unmarshal(res.Body, &result); err != nil {
		return nil, sdkerrors.NewInternal("midaz", operation, "failed to unmarshal response body", err)
	}

	return &result, nil
}

// CreateInflow creates a new inflow transaction (no source account needed).
func (s *transactionsService) CreateInflow(ctx context.Context, orgID, ledgerID string, input *CreateTransactionInflowInput) (*Transaction, error) {
	const operation = "Transactions.CreateInflow"

	if err := ensureService(s); err != nil {
		return nil, err
	}

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "ledger id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "input is required")
	}

	if strings.TrimSpace(input.Send.Asset) == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "send asset is required")
	}

	if strings.TrimSpace(input.Send.Value) == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "send value is required")
	}

	if err := validateTransactionVariantLegs(operation, "distribute", input.Send.Distribute.To); err != nil {
		return nil, err
	}

	return core.Create[Transaction, CreateTransactionInflowInput](ctx, &s.BaseService, transactionsBasePath(orgID, ledgerID)+"/inflow", input)
}

// CreateOutflow creates a new outflow transaction (no destination needed).
func (s *transactionsService) CreateOutflow(ctx context.Context, orgID, ledgerID string, input *CreateTransactionOutflowInput) (*Transaction, error) {
	const operation = "Transactions.CreateOutflow"

	if err := ensureService(s); err != nil {
		return nil, err
	}

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "ledger id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "input is required")
	}

	if strings.TrimSpace(input.Send.Asset) == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "send asset is required")
	}

	if strings.TrimSpace(input.Send.Value) == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "send value is required")
	}

	if err := validateTransactionVariantLegs(operation, "source", input.Send.Source.From); err != nil {
		return nil, err
	}

	return core.Create[Transaction, CreateTransactionOutflowInput](ctx, &s.BaseService, transactionsBasePath(orgID, ledgerID)+"/outflow", input)
}

// Get retrieves a transaction by its unique identifier.
func (s *transactionsService) Get(ctx context.Context, orgID, ledgerID, id string) (*Transaction, error) {
	const operation = "Transactions.Get"

	if err := ensureService(s); err != nil {
		return nil, err
	}

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "ledger id is required")
	}

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "transaction id is required")
	}

	return core.Get[Transaction](ctx, &s.BaseService, transactionsItemPath(orgID, ledgerID, id))
}

// List returns a paginated iterator over transactions in a ledger.
func (s *transactionsService) List(ctx context.Context, orgID, ledgerID string, opts *models.CursorListOptions) *pagination.Iterator[Transaction] {
	if err := ensureService(s); err != nil {
		return newErrorIterator[Transaction](err)
	}

	if orgID == "" || ledgerID == "" {
		return newErrorIterator[Transaction](sdkerrors.NewValidation("Transactions.List", transactionResource, "organization ID and ledger ID are required"))
	}

	return core.List[Transaction](ctx, &s.BaseService, transactionsBasePath(orgID, ledgerID), opts)
}

// Update partially updates an existing transaction.
func (s *transactionsService) Update(ctx context.Context, orgID, ledgerID, id string, input *UpdateTransactionInput) (*Transaction, error) {
	const operation = "Transactions.Update"

	if err := ensureService(s); err != nil {
		return nil, err
	}

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "ledger id is required")
	}

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "transaction id is required")
	}

	if input == nil {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "input is required")
	}

	return core.Update[Transaction, UpdateTransactionInput](ctx, &s.BaseService, transactionsItemPath(orgID, ledgerID, id), input)
}

// Commit transitions the transaction to committed state.
func (s *transactionsService) Commit(ctx context.Context, orgID, ledgerID, id string) (*Transaction, error) {
	const operation = "Transactions.Commit"

	if err := ensureService(s); err != nil {
		return nil, err
	}

	if err := ensureService(s); err != nil {
		return nil, err
	}

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "ledger id is required")
	}

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "transaction id is required")
	}

	return core.Action[Transaction](ctx, &s.BaseService, transactionsItemPath(orgID, ledgerID, id)+"/commit", nil)
}

// Cancel transitions the transaction to cancelled state.
func (s *transactionsService) Cancel(ctx context.Context, orgID, ledgerID, id string) (*Transaction, error) {
	const operation = "Transactions.Cancel"

	if err := ensureService(s); err != nil {
		return nil, err
	}

	if err := ensureService(s); err != nil {
		return nil, err
	}

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "ledger id is required")
	}

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "transaction id is required")
	}

	return core.Action[Transaction](ctx, &s.BaseService, transactionsItemPath(orgID, ledgerID, id)+"/cancel", nil)
}

// Revert creates a reversal of a previously committed transaction.
func (s *transactionsService) Revert(ctx context.Context, orgID, ledgerID, id string) (*Transaction, error) {
	const operation = "Transactions.Revert"

	if err := ensureService(s); err != nil {
		return nil, err
	}

	if err := ensureService(s); err != nil {
		return nil, err
	}

	if orgID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "organization id is required")
	}

	if ledgerID == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "ledger id is required")
	}

	if id == "" {
		return nil, sdkerrors.NewValidation(operation, transactionResource, "transaction id is required")
	}

	return core.Action[Transaction](ctx, &s.BaseService, transactionsItemPath(orgID, ledgerID, id)+"/revert", nil)
}
