# Katapult Certificate Manager Tool

This is a tool which allows you to consume certificates managed by the Katapult Certificate Manager (KCM) service. KCM handles issuing and renewing certificates and using them throughout the Katapult platform. If you want to use those certificates elsewhere, this tool may help you.

Once you have got a certificate in KCM, you can use this tool to monitor that certificate, download the signed certificate and then run any commands you want to run to have those certificates picked up by whatever service is using them.

## Example configuration

By default, configuration lives in `/etc/kcm.yaml` but you can put it wherever you want and use the `--config` flag to specify the path when running `kcm-tool`.

```yaml
certificates:
  - url: https://certs.katapult.io/{certificate-id}/{your-certificate-secret}
    paths:
      certificate: /etc/certs/service.cert.pem
      private_key: /etc/certs/service.key.pem
      chain: /etc/certs/service.chain.pem
      certificate_with_chain: /etc/certs/service.cert-with-chain.pem
    commands:
      - systemctl reload apache2
      - touch /etc/certs/service.cert.updated
```

* You can obtain the `url` attribute through the Katapult interface or API.
* The paths define the paths on your server where you want to export your certificate data.
* The commands will be run in the order provided.

## Getting started

1. Download the latest release from GitHub and pop it in `/usr/local/bin/kcm-tool` on the server.
2. Set the executable flag (`chmod +x /usr/local/bin/kcm-tool`).
3. Add your configuration file somewhere.
4. Run the tool (`kcm-tool --config path/to/config.yaml`).
