# Flash
[![Maintainability Rating](https://sonarqube.redbird.no/api/project_badges/measure?project=flash&metric=sqale_rating&token=sqb_1e44651985a36ed81c0632a28cc0b7259e784780)](https://sonarqube.redbird.no/dashboard?id=flash)
[![Reliability Rating](https://sonarqube.redbird.no/api/project_badges/measure?project=flash&metric=reliability_rating&token=sqb_1e44651985a36ed81c0632a28cc0b7259e784780)](https://sonarqube.redbird.no/dashboard?id=flash)
[![Security Rating](https://sonarqube.redbird.no/api/project_badges/measure?project=flash&metric=security_rating&token=sqb_1e44651985a36ed81c0632a28cc0b7259e784780)](https://sonarqube.redbird.no/dashboard?id=flash)

A TUI based LND node management tool. 

Currently under development. Stay tuned for updates

## Usage

### Authentication ###
Flash uses a unique authentication mechanism that removes the need for storing credentials in cleartext on disk. To set it up you first need to create an encrypted authentication file.

```
./flash -m <admin macaroon file> -c <tls cert file>
```

This will produce an authentication file `auth.bin` and the encryption key will be printed out. 

You can now remove the macaroon and tls file and run flash.

```
./flash -a auth.bin -k 08f89492cc0d12640a580a30747970652e676f6f676c65617069732e636f6d2f676f6f676c652e63727970746f2e74696e6b2e41657347636d4b657912221a20a7c7e86e351fdf1014d2d807d5e3c1db962c91224f7fe4831a9c8717ad412d193801100118f89492cc0ae001
```
