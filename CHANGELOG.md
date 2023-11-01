<a name="Configuration Go Mod Changelog"></a>

## Configuration Module (in Go)
[Github repository](https://github.com/edgexfoundry/go-mod-configuration)

## [v3.1.0] - 2023-11-15

### 👷 Build

- Upgrade to go 1.21 and linter 1.54.2 ([df336b0…](https://github.com/edgexfoundry/go-mod-configuration/commit/df336b0784972208187fc3e6b1297ec5f902d26e))


## [v3.0.0] - 2023-05-31

### Features ✨

- Add GetConfigurationKeys API to return list of keys for config path ([#e857305](https://github.com/edgexfoundry/go-mod-configuration/commits/e857305))
- Update config interface to accept full path ([#23be367](https://github.com/edgexfoundry/go-mod-configuration/commits/23be367))
- Add mocks for configuration ([#dea145b](https://github.com/edgexfoundry/go-mod-configuration/commits/dea145b))

### Code Refactoring ♻

- Remove use of TOML package ([#5ececa](https://github.com/edgexfoundry/go-mod-configuration/commit/5ececa60164dd36dd94c4f9ac90d8d3a341a7359))
  ```text
  BREAKING CHANGE: PutConfigurationToml has been renamed/reworked to be PutConfigurationMap
  ```
- Update module to v3 ([#1331ec2](https://github.com/edgexfoundry/go-mod-configuration/commit/1331ec2abf995885ddb2d2fa53484b1d8dcb7c5a))
  ```text
  BREAKING CHANGE: Import paths will need to change to v3
  ```

### Build 👷

- Update to Go 1.20 and linter v1.51.2 ([#62555dd](https://github.com/edgexfoundry/go-mod-configuration/commits/62555dd))

## [v2.3.0] - 2022-11-09

### Features ✨

- Add the new "Optional" field in ServiceConfig struct ([#c575dfa](https://github.com/edgexfoundry/go-mod-configuration/commits/c575dfa))

### Build 👷

- Upgrade to Go 1.18 ([#26e3452](https://github.com/edgexfoundry/go-mod-configuration/commits/26e3452))

## [v2.2.0] - 2022-05-11

### Features ✨

- None

### Build 👷

- **security:** Enable gosec and default linter set ([#67a3dcd](https://github.com/edgexfoundry/go-mod-configuration/commits/67a3dcd))

## [v2.1.0] - 2021-11-17

### Features ✨

- Add Renew Access Token capability ([#4c2283e](https://github.com/edgexfoundry/go-mod-configuration/commits/4c2283e))

## [v2.0.0] - 2021-06-30

### Features ✨
- **configuration:** Add new HasSubConfiguration(name string) API ([#d38a7b9](https://github.com/edgexfoundry/go-mod-configuration/commits/d38a7b9))
- **security:** add support for ACL access token ([#94e2211](https://github.com/edgexfoundry/go-mod-configuration/commits/94e2211))

<a name="v0.0.3"></a>
## [v0.0.3] - 2020-03-18
### Features ✨
- **configuration:** Add ability for protocol to be defaulted if missing from url spec ([#caf9e31](https://github.com/edgexfoundry/go-mod-configuration/commits/caf9e31))

<a name="v0.0.0"></a>
## v0.0.0 - 2020-01-16
### Features ✨
- **configuration:** Initial abstract Configuration implementation ([#f29fadd](https://github.com/edgexfoundry/go-mod-configuration/commits/f29fadd))
