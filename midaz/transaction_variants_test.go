package midaz

import (
	"bytes"
	"context"
	"errors"
	"mime"
	"mime/multipart"
	"testing"

	"github.com/LerianStudio/lerian-sdk-golang/models"
	sdkerrors "github.com/LerianStudio/lerian-sdk-golang/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionsCreateAnnotation(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/transactions/annotation", path)

			input, ok := body.(*createTransactionRequest)
			require.True(t, ok)
			require.NotNil(t, input.Send)
			assert.Equal(t, "BRL", input.Send.Asset)
			assert.Equal(t, "1.00", input.Send.Value)

			return unmarshalInto(Transaction{ID: "txn-annotation"}, result)
		},
	}

	svc := newTransactionsService(mock)
	txn, err := svc.CreateAnnotation(context.Background(), testOrgID, testLedgerID, &CreateTransactionInput{
		Send: &TransactionSend{
			Asset:      "BRL",
			Value:      "1.00",
			Source:     TransactionSendSource{From: []TransactionOperationLeg{{AccountAlias: "acc-1"}}},
			Distribute: TransactionSendDistribution{To: []TransactionOperationLeg{{AccountAlias: "acc-2"}}},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, txn)
	assert.Equal(t, "txn-annotation", txn.ID)
}

func TestTransactionsCreateDSL(t *testing.T) {
	t.Parallel()

	payload := []byte("SEND 100 BRL")
	mock := &mockBackend{
		callWithHdrsFn: func(_ context.Context, method, path string, headers map[string]string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/transactions/dsl", path)

			raw, ok := body.([]byte)
			require.True(t, ok)

			contentType := headers["Content-Type"]
			mediaType, params, err := mime.ParseMediaType(contentType)
			require.NoError(t, err)
			assert.Equal(t, "multipart/form-data", mediaType)

			reader := multipart.NewReader(bytes.NewReader(raw), params["boundary"])
			part, err := reader.NextPart()
			require.NoError(t, err)
			assert.Equal(t, "transaction", part.FormName())

			partData := new(bytes.Buffer)
			_, err = partData.ReadFrom(part)
			require.NoError(t, err)
			assert.Equal(t, string(payload), partData.String())

			return unmarshalInto(Transaction{ID: "txn-dsl"}, result)
		},
	}

	svc := newTransactionsService(mock)
	txn, err := svc.CreateDSL(context.Background(), testOrgID, testLedgerID, payload)

	require.NoError(t, err)
	require.NotNil(t, txn)
	assert.Equal(t, "txn-dsl", txn.ID)
}

func TestTransactionsCreateDSLEmptyInput(t *testing.T) {
	t.Parallel()

	svc := newTransactionsService(&mockBackend{})
	txn, err := svc.CreateDSL(context.Background(), testOrgID, testLedgerID, nil)

	require.Error(t, err)
	assert.Nil(t, txn)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
}

func TestTransactionsCreateDSLTooLargeInput(t *testing.T) {
	t.Parallel()

	svc := newTransactionsService(&mockBackend{})
	txn, err := svc.CreateDSL(context.Background(), testOrgID, testLedgerID, bytes.Repeat([]byte("x"), maxDSLContentSize+1))

	require.Error(t, err)
	assert.Nil(t, txn)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	assert.Contains(t, err.Error(), "exceeds maximum allowed size")
}

func TestTransactionsCreateInflowValidatesSendShape(t *testing.T) {
	t.Parallel()

	svc := newTransactionsService(&mockBackend{})
	txn, err := svc.CreateInflow(context.Background(), testOrgID, testLedgerID, &CreateTransactionInflowInput{})

	require.Error(t, err)
	assert.Nil(t, txn)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	assert.Contains(t, err.Error(), "send asset is required")

	txn, err = svc.CreateInflow(context.Background(), testOrgID, testLedgerID, &CreateTransactionInflowInput{
		Send: TransactionInflowSend{Asset: "BRL", Value: "1.00"},
	})
	require.Error(t, err)
	assert.Nil(t, txn)
	assert.Contains(t, err.Error(), "distribute legs are required")

	txn, err = svc.CreateInflow(context.Background(), testOrgID, testLedgerID, &CreateTransactionInflowInput{
		Send: TransactionInflowSend{
			Asset:      "BRL",
			Value:      "1.00",
			Distribute: TransactionInflowDistribution{To: []TransactionOperationLeg{{}}},
		},
	})
	require.Error(t, err)
	assert.Nil(t, txn)
	assert.Contains(t, err.Error(), "distribute leg identifier is required")
}

func TestTransactionsCreateOutflowValidatesSendShape(t *testing.T) {
	t.Parallel()

	svc := newTransactionsService(&mockBackend{})
	txn, err := svc.CreateOutflow(context.Background(), testOrgID, testLedgerID, &CreateTransactionOutflowInput{})

	require.Error(t, err)
	assert.Nil(t, txn)
	assert.True(t, errors.Is(err, sdkerrors.ErrValidation))
	assert.Contains(t, err.Error(), "send asset is required")

	txn, err = svc.CreateOutflow(context.Background(), testOrgID, testLedgerID, &CreateTransactionOutflowInput{
		Send: TransactionOutflowSend{Asset: "BRL", Value: "1.00"},
	})
	require.Error(t, err)
	assert.Nil(t, txn)
	assert.Contains(t, err.Error(), "source legs are required")

	txn, err = svc.CreateOutflow(context.Background(), testOrgID, testLedgerID, &CreateTransactionOutflowInput{
		Send: TransactionOutflowSend{
			Asset:  "BRL",
			Value:  "1.00",
			Source: TransactionOutflowSource{From: []TransactionOperationLeg{{}}},
		},
	})
	require.Error(t, err)
	assert.Nil(t, txn)
	assert.Contains(t, err.Error(), "source leg identifier is required")
}

func TestTransactionsCreateInflow(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/transactions/inflow", path)

			input, ok := body.(*CreateTransactionInflowInput)
			require.True(t, ok)
			assert.Equal(t, "BRL", input.Send.Asset)
			require.NotNil(t, input.Route)
			assert.Equal(t, "route-inflow", *input.Route)
			assert.Len(t, input.Send.Distribute.To, 1)
			assert.Equal(t, "freeze", input.Send.Distribute.To[0].BalanceKey)
			assert.Equal(t, "leg-route", input.Send.Distribute.To[0].Route)

			return unmarshalInto(Transaction{ID: "txn-inflow", Status: models.Status{Code: "pending"}}, result)
		},
	}

	svc := newTransactionsService(mock)
	route := "route-inflow"
	txn, err := svc.CreateInflow(context.Background(), testOrgID, testLedgerID, &CreateTransactionInflowInput{
		Route: &route,
		Send: TransactionInflowSend{
			Asset: "BRL",
			Value: "100.00",
			Distribute: TransactionInflowDistribution{To: []TransactionOperationLeg{{
				AccountAlias: "@customer",
				BalanceKey:   "freeze",
				Route:        "leg-route",
			}}},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, txn)
	assert.Equal(t, "txn-inflow", txn.ID)
}

func TestTransactionsCreateOutflow(t *testing.T) {
	t.Parallel()

	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "POST", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/transactions/outflow", path)

			input, ok := body.(*CreateTransactionOutflowInput)
			require.True(t, ok)
			assert.Equal(t, "BRL", input.Send.Asset)
			require.NotNil(t, input.Route)
			assert.Equal(t, "route-outflow", *input.Route)
			assert.Len(t, input.Send.Source.From, 1)
			assert.Equal(t, "freeze", input.Send.Source.From[0].BalanceKey)
			assert.Equal(t, "leg-route", input.Send.Source.From[0].Route)

			return unmarshalInto(Transaction{ID: "txn-outflow", Status: models.Status{Code: "pending"}}, result)
		},
	}

	svc := newTransactionsService(mock)
	pending := true
	route := "route-outflow"
	txn, err := svc.CreateOutflow(context.Background(), testOrgID, testLedgerID, &CreateTransactionOutflowInput{
		Pending: &pending,
		Route:   &route,
		Send: TransactionOutflowSend{
			Asset: "BRL",
			Value: "100.00",
			Source: TransactionOutflowSource{From: []TransactionOperationLeg{{
				AccountAlias: "@origin",
				BalanceKey:   "freeze",
				Route:        "leg-route",
			}}},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, txn)
	assert.Equal(t, "txn-outflow", txn.ID)
}

func TestOperationsUpdate(t *testing.T) {
	t.Parallel()

	description := "updated operation"
	mock := &mockBackend{
		callFn: func(_ context.Context, method, path string, body, result any) error {
			assert.Equal(t, "PATCH", method)
			assert.Equal(t, "/organizations/org-1/ledgers/led-1/transactions/txn-1/operations/op-1", path)

			input, ok := body.(*UpdateOperationInput)
			require.True(t, ok)
			require.NotNil(t, input.Description)
			assert.Equal(t, description, *input.Description)

			return unmarshalInto(Operation{ID: "op-1", TransactionID: "txn-1", Description: &description}, result)
		},
	}

	svc := newOperationsService(mock)
	op, err := svc.Update(context.Background(), testOrgID, testLedgerID, "txn-1", "op-1", &UpdateOperationInput{Description: &description})

	require.NoError(t, err)
	require.NotNil(t, op)
	assert.Equal(t, "op-1", op.ID)
	require.NotNil(t, op.Description)
	assert.Equal(t, description, *op.Description)
}
