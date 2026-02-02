# hubspot-cli

A command-line interface for HubSpot CRM. Manage contacts, companies, deals, and more from your terminal.

## Features

- **CRM Objects** - Full CRUD operations for contacts, companies, deals, tickets, products, quotes, and line items
- **Engagements** - Manage notes, calls, emails, meetings, and tasks
- **Marketing** - Access forms, campaigns, and marketing emails
- **CMS** - Manage files, pages, blogs, and HubDB tables
- **Automation** - List and manage workflows, enroll objects
- **GraphQL** - Execute queries and explore the schema
- **Multiple output formats** - Table (default), JSON, or plain text

## Installation

### macOS

**Homebrew (recommended)**

```bash
brew install open-cli-collective/tap/hspt
```

---

### Windows

**Chocolatey**

```powershell
choco install hspt
```

**Winget**

```powershell
winget install OpenCLICollective.hspt
```

---

### Linux

**Binary download**

Download `.deb`, `.rpm`, or `.tar.gz` from the [Releases page](https://github.com/open-cli-collective/hubspot-cli/releases).

---

### From Source

```bash
go install github.com/open-cli-collective/hubspot-cli/cmd/hspt@latest
```

## Setup

### 1. Create a HubSpot Private App

1. Go to your HubSpot account **Settings** > **Integrations** > **Private Apps**
2. Click **Create a private app**
3. Name it (e.g., "CLI Access")
4. Under **Scopes**, select the permissions you need:
   - CRM: `crm.objects.contacts.read`, `crm.objects.contacts.write`, etc.
   - Marketing: `forms`, `content`
   - Automation: `automation`
5. Click **Create app**
6. Copy the access token shown

### 2. Configure hspt

Run the init command:

```bash
hspt init
```

You'll be prompted to enter your access token. Alternatively, set it via environment variable:

```bash
export HUBSPOT_ACCESS_TOKEN=pat-na1-xxxxx
```

### 3. Verify Setup

```bash
hspt config test
```

## Commands

### Configuration

```bash
# Guided setup
hspt init

# Check configuration status
hspt config show

# Set access token
hspt config set --token pat-na1-xxxxx

# Test API connectivity
hspt config test

# Clear stored configuration
hspt config clear

# Show version
hspt --version
```

### CRM Objects

All CRM object commands follow the same pattern with `list`, `get`, `create`, `update`, `delete`, and `search` subcommands.

| Command | Description |
|---------|-------------|
| `contacts` | Manage CRM contacts |
| `companies` | Manage CRM companies |
| `deals` | Manage CRM deals |
| `tickets` | Manage support tickets |
| `owners` | View CRM owners (users) |
| `products` | Manage products |
| `line-items` | Manage line items |
| `quotes` | Manage quotes |

**Examples:**

```bash
# List contacts
hspt contacts list --limit 20

# Search contacts by email
hspt contacts search --email john@example.com

# Get a specific contact
hspt contacts get 12345

# Create a contact
hspt contacts create --email jane@example.com --firstname Jane --lastname Doe

# Update a contact
hspt contacts update 12345 --phone "+1-555-0100"

# Delete a contact (requires --force)
hspt contacts delete 12345 --force

# Output as JSON
hspt contacts list -o json
```

```bash
# List deals with custom properties
hspt deals list --properties dealname,amount,closedate

# Search deals by stage
hspt deals search --stage closedwon --limit 50

# Create a deal
hspt deals create --name "Enterprise Deal" --amount 50000 --stage qualifiedtobuy
```

### Engagements

| Command | Description |
|---------|-------------|
| `notes` | Manage notes attached to records |
| `calls` | Manage call records |
| `emails` | Manage email activities |
| `meetings` | Manage meeting records |
| `tasks` | Manage tasks |

**Examples:**

```bash
# List recent notes
hspt notes list --limit 10

# Create a task
hspt tasks create --subject "Follow up" --body "Call about renewal" --priority HIGH

# Log a call
hspt calls create --body "Discussed pricing" --direction OUTBOUND --duration 300
```

### Associations

```bash
# List companies associated with a contact
hspt associations list --from-type contacts --from-id 123 --to-type companies

# Create an association
hspt associations create --from-type contacts --from-id 123 --to-type companies --to-id 456

# Delete an association (requires --force)
hspt associations delete --from-type contacts --from-id 123 --to-type companies --to-id 456 --force
```

### Properties

```bash
# List properties for contacts
hspt properties list --object-type contacts

# Get property details
hspt properties get --object-type contacts --name email

# Create a custom property
hspt properties create --object-type contacts --name custom_field --label "Custom Field" --type string --field-type text

# Delete a property (requires --force)
hspt properties delete --object-type contacts --name custom_field --force
```

### Pipelines

```bash
# List deal pipelines
hspt pipelines list --object-type deals

# Get pipeline details
hspt pipelines get --object-type deals --id default

# List pipeline stages
hspt pipelines stages --object-type deals --id default
```

### Custom Object Schemas

Requires Operations Hub Professional or Enterprise.

```bash
# List custom object schemas
hspt schemas list

# Get schema details
hspt schemas get p_my_custom_object

# Create schema from JSON file
hspt schemas create --file schema.json

# Delete schema (requires --force)
hspt schemas delete p_my_custom_object --force
```

### Marketing

| Command | Description |
|---------|-------------|
| `forms` | View forms and submissions |
| `campaigns` | View marketing campaigns |
| `marketing-emails` | Manage marketing emails |

**Examples:**

```bash
# List forms
hspt forms list

# Get form submissions
hspt forms submissions <form-id>

# List campaigns
hspt campaigns list

# List marketing emails
hspt marketing-emails list
```

### CMS

| Command | Description |
|---------|-------------|
| `files` | Manage files in File Manager |
| `domains` | View domains |
| `pages` | Manage site and landing pages |
| `blogs` | Manage blog posts, authors, and tags |
| `hubdb` | Manage HubDB tables and rows |

**Examples:**

```bash
# List files
hspt files list

# List folders
hspt files folders

# List site pages
hspt pages list --type site

# List landing pages
hspt pages list --type landing
```

**HubDB:**

```bash
# List HubDB tables
hspt hubdb tables list

# Get table details
hspt hubdb tables get my_table

# List rows in a table
hspt hubdb rows list my_table

# Create a row
hspt hubdb rows create my_table --file row.json

# Publish table changes
hspt hubdb tables publish my_table
```

### Conversations

```bash
# List inboxes
hspt conversations inboxes list

# List threads
hspt conversations threads list

# Get thread details
hspt conversations threads get <thread-id>

# List messages in a thread
hspt conversations messages list <thread-id>

# Send a message
hspt conversations messages send <thread-id> --text "Hello!" --channel-id <id>
```

### Workflows

```bash
# List workflows
hspt workflows list

# Get workflow details
hspt workflows get <workflow-id>

# Create workflow from JSON
hspt workflows create --file workflow.json

# Update workflow
hspt workflows update <workflow-id> --file workflow.json

# Delete workflow (requires --force)
hspt workflows delete <workflow-id> --force

# Enroll an object in a workflow
hspt workflows enroll <workflow-id> --object-id <contact-id>

# List workflow enrollments
hspt workflows enrollments <workflow-id>
```

### GraphQL

Execute queries against HubSpot's unified GraphQL API.

```bash
# Execute query from file
hspt graphql query --file query.graphql

# Execute inline query
hspt graphql query --query '{ CRM { contact_collection(limit: 10) { items { email } } } }'

# Query with associations
hspt graphql query --query '{
  CRM {
    contact_collection(limit: 5) {
      items {
        firstname
        email
        associations {
          company_collection { items { name } }
        }
      }
    }
  }
}'
```

**Schema exploration:**

```bash
# List available types
hspt graphql explore

# Show fields of a type
hspt graphql explore --type CRM

# Show field details
hspt graphql explore --type CRM --field contact_collection
```

## Global Flags

All commands support these flags:

| Flag | Description |
|------|-------------|
| `-o, --output` | Output format: `table` (default), `json`, `plain` |
| `--no-color` | Disable colored output |
| `-v, --verbose` | Enable verbose output |

**Examples:**

```bash
# JSON output for scripting
hspt contacts list -o json

# Plain output for simple lists
hspt contacts list -o plain

# Disable colors (useful in CI)
hspt contacts list --no-color
```

## Common Patterns

### Pagination

Most list commands support pagination:

```bash
# Limit results
hspt contacts list --limit 50

# Get next page
hspt contacts list --after <cursor>
```

The cursor for the next page is shown in the output when more results are available.

### Custom Properties

Specify which properties to return:

```bash
hspt contacts list --properties email,firstname,lastname,company
```

Set custom properties when creating or updating:

```bash
hspt contacts create --email test@example.com --prop custom_field=value --prop another_field=123
```

### Destructive Operations

Delete commands require `--force` to confirm:

```bash
# Shows warning without executing
hspt contacts delete 12345

# Actually deletes
hspt contacts delete 12345 --force
```

## Shell Completion

hspt supports tab completion for bash, zsh, fish, and PowerShell.

### Bash

```bash
# Load in current session
source <(hspt completion bash)

# Install permanently (Linux)
hspt completion bash | sudo tee /etc/bash_completion.d/hspt > /dev/null

# Install permanently (macOS with Homebrew)
hspt completion bash > $(brew --prefix)/etc/bash_completion.d/hspt
```

### Zsh

```bash
# Load in current session
source <(hspt completion zsh)

# Install permanently
mkdir -p ~/.zsh/completions
hspt completion zsh > ~/.zsh/completions/_hspt

# Add to ~/.zshrc if not already present:
# fpath=(~/.zsh/completions $fpath)
# autoload -Uz compinit && compinit
```

### Fish

```bash
# Load in current session
hspt completion fish | source

# Install permanently
hspt completion fish > ~/.config/fish/completions/hspt.fish
```

### PowerShell

```powershell
# Load in current session
hspt completion powershell | Out-String | Invoke-Expression

# Install permanently (add to $PROFILE)
hspt completion powershell >> $PROFILE
```

## Configuration

Configuration is stored in `~/.config/hubspot-cli/config.json`:

```json
{
  "access_token": "pat-na1-xxxxx"
}
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `HUBSPOT_ACCESS_TOKEN` | HubSpot private app access token |

Environment variables take precedence over the config file.

## Troubleshooting

### "401 Unauthorized" or "Invalid token"

Your access token is invalid or expired:

```bash
hspt config clear
hspt init
```

### "403 Forbidden" or "Missing scopes"

Your private app doesn't have the required scopes for this operation. Edit your private app in HubSpot Settings to add the needed scopes.

### Test Connection

Verify your configuration:

```bash
hspt config test
```

### Verbose Mode

For debugging, use verbose mode:

```bash
hspt --verbose contacts list
```

## License

MIT License - see [LICENSE](LICENSE) for details.
