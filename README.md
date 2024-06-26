# Witness Webhook

Witness Webhook is a service that will listen for webhooks and create signed attestations from them.

## Configuring

Witness Webhook is configured primarily through a YAML file, with a few environment variables.

### Environment Variables

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `WITNESS_WEBHOOK_CONFIG_PATH` | `/webhook-config.yaml` | The path to the yaml config file |
| `WITNESS_WEBHOOK_LISTEN_ADDR` | `:8085` | The address the server will listen on |

### YAML Config

A sample config file follows:

```yaml
archivistaUrl: https://archivista.localhost
attestationDirectory: /tmp
webhooks:
  webhook1:
    type: github
    options:
      secret-file-path: /githubwebhooksecret
    signer: kms
    signerOptions:
      ref: hashivault://testkey
      hashivault-addr: https://somevaultinstance:8082
```

This config defines one webhook named `webhook1`. The server will listen for webhooks from Github at `http://host:8085/webhook/webhook1`. Attestations from this webhook will be signed by a KMS signer that points to a Vault transit key named `testkey`.

Signed attestations will be pushed to an Archivista instance at `https://archivista.localhost`. They will also be stored on disk at `/tmp`. If `archivistaUrl` or `attestationDirectory` are left blank, witness-webhook will not attempt to output attestations to the respective destinations.

For a list of available signers, take a look at the signers provided by [go-witness](https://github.com/in-toto/go-witness). All options in `signerOptions` will be forwarded to the specified signer provider. This section will be updated to include the available signers and options soon.

Currently the available webhook handlers are:

#### Github

| Option | Default Value | Description |
| ------ | ------------- | ----------- |
| `secret-file-path` | | Path to the file containing the secret used to verify webhooks from Github |