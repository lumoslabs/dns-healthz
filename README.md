# DNS Healthz

Just a simple little program to check for good responses from Domain Name Servers.

## Usage

usage: dns-healthz [<flags>]

    A DNS healthz server.

    Flags:
      --help               Show context-sensitive help (also try --help-long and
                           --help-man).
    -C, --config=CONFIG      Path to config file.
      --v=0                Enable V-leveled logging at the specified level.
      --vmodule=""         The syntax of the argument is a comma-separated list
                           of pattern=N,
                             where pattern is a literal file name (minus the ".go" suffix) or
                             "glob" pattern and N is a V level. For instance,
                              -vmodule=gopher*=3
                             sets the V level to 3 in all Go files whose names begin "gopher".
      --probe=PROBE ...    DNS probe as json string
    -l, --listen=":8080"     Listen address
      --grace-timeout=30s  Graceful shutdown timeout.
      --version            Show application version.

## Configuration

Either define probes on the command line or in a config file. See [example.yml] for config file syntax.
