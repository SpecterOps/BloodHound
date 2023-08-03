# Exercising SAML Flows with Test IDP

The BloodHound API comes with a basic SAML IDP implementation to test SAML SSO login flows:

```text
$ just build testidp
$ ./dist/bin/testidp

Usage of ./testidp:
  -bind string
        Address to bind on. If this value has a colon, as in ":8000" or
                        "127.0.0.1:9001", it will be treated as a TCP address. If it
                        begins with a "/" or a ".", it will be treated as a path to a
                        UNIX socket. If it begins with the string "fd@", as in "fd@3",
                        it will be treated as a file descriptor (useful for use with
                        systemd, for instance). If it begins with the string "einhorn@",
                        as in "einhorn@0", the corresponding einhorn socket will be
                        used. If an option is not explicitly passed, the implementation
                        will automatically select among "einhorn@0" (Einhorn), "fd@3"
                        (systemd), and ":8000" (fallback) based on its environment. (default ":8000")
  -cert string
        Path to the IDP cert. If the file doesn't exist a new cert and keypair will be created.
  -key string
        Path to the IDP key. If the file doesn't exist a new cert and keypair will be created.
  -org string
        The organization of the IDP. (default "example.com")
  -url string
        The base URL to the IDP.
```

## Starting the Test IDP

The test IDP server is simple to initialize. It will create a new x509 certificate and RSA private key if neither exist
at the locations provided.

```bash
$ just build testidp
$ ./dist/bin/testidp org onetruesso.com -bind localhost:8081 -url http://localhost:8081 -key ~/testidp.key -cert ~/testidp.cert
```

The test IDP server hosts the several endpoints with the following being the most critical to SSO login:

* `http://localhost:8081/sso` The SSO initiation endpoint. This serves a simple form that a user can then attempt to
  login with.
* `http://localhost:8081/metadata` The IDP metadata description endpoint. This serves all relevant information about the
  IDP including the signing and encryption keys that the IDP employs.

### Default IDP Users

The test IDP server comes with several pre-generated users that can exercise SSO login flows:

| Username | Password |
| --- | --- |
| alice | hunter2 |
| bob | hunter2 |
