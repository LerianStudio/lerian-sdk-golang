## [1.0.0-beta.2](https://github.com/LerianStudio/lerian-sdk-golang/compare/v1.0.0-beta.1...v1.0.0-beta.2) (2026-03-25)


### Features

* **midaz:** add CRM services for holders and aliases ([85af54c](https://github.com/LerianStudio/lerian-sdk-golang/commit/85af54c55b3cb668dab99e4d92ef098ec41b07e9))
* **pkg/pagination:** add numbered-page iterator adapter for CRM pagination ([2e1b79c](https://github.com/LerianStudio/lerian-sdk-golang/commit/2e1b79ce6c30596e81984f4cfff6b518f526fd39))

## [0.2.0](https://github.com/LerianStudio/lerian-sdk-golang/compare/v0.1.0...v0.2.0) (2026-03-25)


### Features

* **midaz:** add CRM services for holders and aliases ([85af54c](https://github.com/LerianStudio/lerian-sdk-golang/commit/85af54c55b3cb668dab99e4d92ef098ec41b07e9))
* **pkg/pagination:** add numbered-page iterator adapter for CRM pagination ([2e1b79c](https://github.com/LerianStudio/lerian-sdk-golang/commit/2e1b79ce6c30596e81984f4cfff6b518f526fd39))

## [0.1.0](https://github.com/LerianStudio/lerian-sdk-golang/compare/v0.0.1...v0.1.0) (2026-03-24)


### Features

* **auth:** adopt Lerian-native token contract ([4bb8e8d](https://github.com/LerianStudio/lerian-sdk-golang/commit/4bb8e8d0c9ec643c2907635741a5470c806b3a19))


### Bug Fixes

* address CodeRabbit review comments ([57854b3](https://github.com/LerianStudio/lerian-sdk-golang/commit/57854b356a15b8928d71c751bc968d11d079ce88))
* **auth:** normalize token redirect checks ([f21cbc4](https://github.com/LerianStudio/lerian-sdk-golang/commit/f21cbc4cb20e67d32dc0cc85fd09fa7ed6a7794b))
* simplify GolangCI-Lint job and fix Trivy action version ([43d5cc8](https://github.com/LerianStudio/lerian-sdk-golang/commit/43d5cc84394e58e76e29470d53a9c964467a832c))
* use golangci-lint v2 directly instead of shared action ([dc27bb8](https://github.com/LerianStudio/lerian-sdk-golang/commit/dc27bb8a1327304fc4f23f2943c716367e772207))

## 1.0.0-beta.1 (2026-03-24)


### Features

* **auth:** adopt Lerian-native token contract ([4bb8e8d](https://github.com/LerianStudio/lerian-sdk-golang/commit/4bb8e8d0c9ec643c2907635741a5470c806b3a19))
* **auth:** migrate from token/APIKey to OAuth2 client credentials ([bf198d7](https://github.com/LerianStudio/lerian-sdk-golang/commit/bf198d74b5af8c3d7fdc0b9079176c322448be88))
* **client:** implement OAuth2 configuration and validation ([347740c](https://github.com/LerianStudio/lerian-sdk-golang/commit/347740ccf3b311da594d3c0510be0623294072d0))
* **core:** add support for tenant ID injection via context ([ba762fc](https://github.com/LerianStudio/lerian-sdk-golang/commit/ba762fc761c0f6a8bf64f228635d511d943ab184))
* **env:** add OAuth2 environment variable support ([a6a13ea](https://github.com/LerianStudio/lerian-sdk-golang/commit/a6a13ea54a977c2677861025900e43d655c04afc))
* **fees:** add DSL transaction transformation support ([7936f1e](https://github.com/LerianStudio/lerian-sdk-golang/commit/7936f1e52b4354866d3baaa7fc32c78331eb86aa))
* **products:** migrate midaz, matcher, tracer, reporter to OAuth2 ([8ad9141](https://github.com/LerianStudio/lerian-sdk-golang/commit/8ad9141a3aca16136ecee195213c32312475a5d2))


### Bug Fixes

* address CodeRabbit review comments ([57854b3](https://github.com/LerianStudio/lerian-sdk-golang/commit/57854b356a15b8928d71c751bc968d11d079ce88))
* **auth:** normalize token redirect checks ([f21cbc4](https://github.com/LerianStudio/lerian-sdk-golang/commit/f21cbc4cb20e67d32dc0cc85fd09fa7ed6a7794b))
* **ci:** resolve lint issues and add make ci target ([84c1a1f](https://github.com/LerianStudio/lerian-sdk-golang/commit/84c1a1f8c037029df771a10682a37de5219aedac))
* **ci:** upgrade golangci-lint to v2.11.4 and add gosec exclusions ([178376a](https://github.com/LerianStudio/lerian-sdk-golang/commit/178376a010e3e3d2d4a0b49bb393d38d64c93d62))
* **deps:** upgrade google.golang.org/grpc to v1.79.3 ([a8f8443](https://github.com/LerianStudio/lerian-sdk-golang/commit/a8f8443a40a936d63ac7e369bcf656b5d1e795a9)), closes [#1](https://github.com/LerianStudio/lerian-sdk-golang/issues/1)
* **docs:** correct Organization type in README and grpc version in PROJECT_RULES ([400ec04](https://github.com/LerianStudio/lerian-sdk-golang/commit/400ec04f0e268f96e2056fc76a0b1323a7bf14a1))
* harden production code with type safety and dead code removal ([d465f68](https://github.com/LerianStudio/lerian-sdk-golang/commit/d465f680209dde1a90d2d087957ff0db4d1399e3))
* simplify GolangCI-Lint job and fix Trivy action version ([43d5cc8](https://github.com/LerianStudio/lerian-sdk-golang/commit/43d5cc84394e58e76e29470d53a9c964467a832c))
* use golangci-lint v2 directly instead of shared action ([dc27bb8](https://github.com/LerianStudio/lerian-sdk-golang/commit/dc27bb8a1327304fc4f23f2943c716367e772207))
