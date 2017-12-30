# Observatory Coordinator Change Log

## Beta Notes
The REST API is in flux and **not** locked down. Stability testing is ongoing
but not yet proven. Resource usage and capacity planning have not yet been
assessed.

### Not Yet Implemented
The following functionality is planned for inclusion in the first stable release,
but has not yet been implemented in the current beta release:
- UI and API
 - UI and API user authentication and authorization
 - Form validation and friendly error messages in the UI
 - UI auto-refresh
 - System configuration via the UI
- Storage & retrieval of raw check output
- Additional period types
- Check result confirmation and flap suppression
- PagerDuty alert integration
- HTML email alerts
- Data expiration - *currently all data lives forever and the database will grow
indefinitely*.

## v0.4.0
- **Alerts now fire on state change instead of every non-OK check result**
- **Alert template variable names have changed**
- **Command-line flags have changed, see `--help`**
- Added configurable reminders to alerts
- CPU, memory, disk, and port listener checks
- Drill down by role on dashboard
- Added system status page to web UI
- Built with Go 1.7 stable

## v0.3.2
- UI endpoint is more configurable
- Basic documentation
- Basic cross-compilation & build packaging
- Test builds wit Go 1.7 release candidates

## v0.3.1
- **Breaking changes will invalidate existing data in MongoDB**
- Switched entity IDs from BSON ObjectIDs to UUIDs
- Changed underlying values for check and alert type enumerations
- Added -samplefill option to add some test checks and alerts to the database
- The Coordinator can serve the UI directly

## v0.3.0
- HTTP checks
- Agent check-in checks
- Quiet & blackout periods
- Email alerts

## v0.2.0
- UI look & feel
- Dashboard
- Subject, check, and alert management

## v0.1.0
- Coordinator peer management
- Agent bootstrapping & configuration updating
- Executable checks & alerts
