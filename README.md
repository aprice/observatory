# Observatory Coordinator

## Installation
The coordinator (or coordinator.exe on Windows) binary is fully self-contained
and can be copied to any appropriate location as desired.

### MongoDB
The coordinator requires access to a MongoDB instance for storing data. For
more information, see the Configuration section below.

## Usage
Some configuration details can be passed on the command line:
- `--help`: display help information
- `--version`: display version information and exit
- `--config`: path to a configuration file (see Configuration section below)
- `--port`: port to run the REST API on (default 13100)
- `--address`: advertised REST API address (default auto-detect)
  - *This is only the address that will be sent to other nodes as the canonical
  address for this node.*
- `--peers`: comma-separated list of other coordinator node endpoints to connect
to when bootstrapping (see Peering section below)

### Configuration
The configuration file can provide many of the same options, and a broad set of
additional settings. *Any settings provided on the command line will override the
equivalent seting in the configuration file.* The configuration is a text file in JSON format:
```javascript
{
	"Port": 13100,
	"Address": "192.168.13.1",
	"AgentUpdateInterval": 15,
	"PeerUpdateInterval": 15,
	"PeerCheckInterval": 5,
	"MongoHost": "localhost",
	"MongoDatabase": "Observatory",
	"BootstrapPeers": [],
	"UIFilePath": "/opt/observeratory/ui",
	"UIWebPath": "ui",
	"UIPort": 8000
}
```

The available settings are:
- `Port`: integer port to expose the REST API on (same as the `-port` flag)
- `Address`: address to advertise (same as the `-address` flag)
- `AgentUpdateInterval`: time (in seconds) between agent check-ins (default `30`)
- `PeerUpdateInterval`: time (in seconds) between querying peers for their peers
to update the peer list (default `30`)
- `PeerCheckInterval`: time (in seconds) between checking if peer coordinators
are up (default `5`)
- `MongoHost`: address for the MongoDB server (default `"localhost"`)
- `MongoDatabase`: the database name to use (default `"Observatory"`)
- `MongoUser`: username to authenticate with MongoDB, if any
- `MongoPassword`: password to authenticate with MongoDB, if any
- `BootstrapPeers`: peers to connect to at startup (same as the `-peers` flag)
- `SMTPHost`: SMTP host address for sending alert e-mails (optional)
- `SMTPPort`: SMTP host port (optional)
- `SMTPUser`: username for authenticating with the SMTP server, if any
- `SMTPPassword`: password for authenticating with the SMTP server, if any
- `EmailFrom`: "from" address to use for alert e-mails (optional)

### Expiring Data
Currently, the record of every executed check is retained indefinitely. You can,
however, control the data set size by adding a ttl index directly in MongoDB:
```javascript
db.CheckResults.createindex({"time" : -1}, {
    "name":"ttl",
    "expireAfterSeconds":604800
})
```

This example uses a retention of 604800 seconds, or 1 week.

# Observatory Agent

## Installation
The agent (or agent.exe on Windows) binary is fully self-contained and can be
copied to any appropriate location as desired.

## Usage
The agent does not use a configuration file. You can pass parameters to the
agent at runtime to provide configuration details:
- `--help`: display help information
- `--version`: display version information and exit
- `--name`: the name to be given to this Subject (defaults to hostname)
- `--roles`: provide initial Roles for bootstrapping a new Subject (comma-separated)
- `--coordinator`: provide an initial Coordinator endpoint (e.g. 192.168.13.1:13100)
  - *Additional coordinators will be automatically detected when the agent starts.*

On startup, the agent will connect to the given coordinator. If this is an agent
for a new Subject, it will automatically register the new Subject with the given
roles and name.

The agent will get all configuration for the pool of coordinators, checks to
execute, and so on from the coordinator when it starts up, and it will update
its configuration on a regular basis.
