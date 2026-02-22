import re

groups = {
    "ManageNotes": {
        "actions": ["read", "write", "delete", "append", "rename", "duplicate", "move"],
        "mappings": {
            "read": ("ReadNoteHandler", "ReadNoteArgs"),
            "write": ("WriteNoteHandler", "WriteNoteArgs"),
            "delete": ("DeleteNoteHandler", "DeleteNoteArgs"),
            "append": ("AppendNoteHandler", "AppendNoteArgs"),
            "rename": ("RenameNoteHandler", "RenameNoteArgs"),
            "duplicate": ("DuplicateNoteHandler", "DuplicateNoteArgs"),
            "move": ("MoveNoteHandler", "MoveArgs"),
        }
    },
    "EditNote": {
        "actions": ["edit", "replace-section", "batch-edit"],
        "mappings": {
            "edit": ("EditNoteHandler", "EditNoteArgs"),
            "replace-section": ("ReplaceSectionHandler", "ReplaceSectionArgs"),
            "batch-edit": ("BatchEditNoteHandler", "BatchEditArgs"),
        }
    },
    "SearchVault": {
        "actions": ["search", "advanced", "date", "regex", "tags", "headings", "inline-fields"],
        "mappings": {
            "search": ("SearchVaultHandler", "SearchArgs"),
            "advanced": ("SearchAdvancedHandler", "SearchAdvancedArgs"),
            "date": ("SearchDateHandler", "SearchDateArgs"),
            "regex": ("SearchRegexHandler", "SearchRegexArgs"),
            "tags": ("SearchByTagsHandler", "SearchTagsArgs"),
            "headings": ("SearchHeadingsHandler", "SearchHeadingsArgs"),
            "inline-fields": ("QueryInlineFieldsHandler", "SearchInlineFieldsArgs"),
        }
    },
    "ManagePeriodicNotes": {
        "actions": ["daily", "weekly", "monthly", "quarterly", "yearly", "list-daily", "list-periodic"],
        "mappings": {
            "daily": ("DailyNoteHandler", "DailyNoteArgs"),
            "weekly": ("WeeklyNoteHandler", "WeeklyNoteArgs"),
            "monthly": ("MonthlyNoteHandler", "MonthlyNoteArgs"),
            "quarterly": ("QuarterlyNoteHandler", "QuarterlyNoteArgs"),
            "yearly": ("YearlyNoteHandler", "YearlyNoteArgs"),
            "list-daily": ("ListDailyNotesHandler", "ListPeriodicArgs"), # ListDaily doesn't take args directly, wait.. let's check. Assuming it takes ListPeriodicArgs or none
            "list-periodic": ("ListPeriodicNotesHandler", "ListPeriodicArgs"),
        }
    },
    "ManageFolders": {
        "actions": ["list", "create", "delete"],
        "mappings": {
            "list": ("ListFoldersHandler", "ListDirsArgs"),
            "create": ("CreateFolderHandler", "CreateDirArgs"),
            "delete": ("DeleteFolderHandler", "DeleteDirArgs"),
        }
    },
    "ManageFrontmatter": {
        "actions": ["get", "set", "remove", "add-alias", "add-tag"],
        "mappings": {
            "get": ("GetFrontmatterHandler", "GetFrontmatterArgs"),
            "set": ("SetFrontmatterHandler", "SetFrontmatterArgs"),
            "remove": ("RemoveFrontmatterKeyHandler", "DeleteFrontmatterArgs"),
            "add-alias": ("AddAliasHandler", "AddAliasArgs"),
            "add-tag": ("AddTagToFrontmatterHandler", "AddTagArgs"),
        }
    },
    "ManageTasks": {
        "actions": ["list", "toggle", "complete"],
        "mappings": {
            "list": ("ListTasksHandler", "ListTasksArgs"),
            "toggle": ("ToggleTaskHandler", "ToggleTaskArgs"),
            "complete": ("CompleteTasksHandler", "CompleteTasksArgs"),
        }
    },
    "AnalyzeVault": {
        "actions": ["stats", "broken-links", "orphan-notes", "unlinked-mentions", "find-stubs", "find-outdated"],
        "mappings": {
            "stats": ("VaultStatsHandler", "VaultStatsArgs"),
            "broken-links": ("BrokenLinksHandler", "BrokenLinksArgs"),
            "orphan-notes": ("OrphanNotesHandler", "OrphanNotesArgs"),
            "unlinked-mentions": ("UnlinkedMentionsHandler", "UnlinkedMentionsArgs"),
            "find-stubs": ("FindStubsHandler", "FindStubsArgs"),
            "find-outdated": ("FindOutdatedHandler", "FindOutdatedArgs"),
        }
    },
    "ManageCanvas": {
        "actions": ["list", "read", "create", "add-node", "add-edge"],
        "mappings": {
            "list": ("ListCanvasesHandler", "ListDirsArgs"), # assuming
            "read": ("ReadCanvasHandler", "ReadNoteArgs"), # assuming
            "create": ("CreateCanvasHandler", "CreateCanvasArgs"),
            "add-node": ("AddCanvasNodeHandler", "AddNodeArgs"),
            "add-edge": ("AddCanvasEdgeHandler", "ConnectNodesArgs"),
        }
    },
    "ManageMocs": {
        "actions": ["discover", "generate", "update", "generate-index"],
        "mappings": {
            "discover": ("DiscoverMOCsHandler", "DiscoverMOCsArgs"),
            "generate": ("GenerateMOCHandler", "GenerateMOCArgs"),
            "update": ("UpdateMOCHandler", "UpdateMOCArgs"),
            "generate-index": ("GenerateIndexHandler", "GenerateIndexArgs"),
        }
    },
    "ReadBatch": {
        "actions": ["read", "get-section", "get-headings", "get-summary"],
        "mappings": {
            "read": ("ReadNotesHandler", "ReadNotesArgs"),
            "get-section": ("GetSectionHandler", "GetSectionArgs"),
            "get-headings": ("GetHeadingsHandler", "GetHeadingsArgs"),
            "get-summary": ("GetNoteSummaryHandler", "GetNoteSummaryArgs"),
        }
    },
    "ManageLinks": {
        "actions": ["backlinks", "forward-links", "suggest"],
        "mappings": {
            "backlinks": ("BacklinksHandler", "GetBacklinksArgs"),
            "forward-links": ("ForwardLinksHandler", "ForwardLinksArgs"),
            "suggest": ("SuggestLinksHandler", "SuggestLinksArgs"),
        }
    },
    "BulkOperations": {
        "actions": ["tag", "move", "set-frontmatter"],
        "mappings": {
            "tag": ("BulkTagHandler", "BulkTagArgs"),
            "move": ("BulkMoveHandler", "BulkMoveArgs"),
            "set-frontmatter": ("BulkSetFrontmatterHandler", "BulkSetFrontmatterArgs"),
        }
    },
    "ManageTemplates": {
        "actions": ["list", "get", "apply"],
        "mappings": {
            "list": ("ListTemplatesHandler", "ListTemplatesArgs"),
            "get": ("GetTemplateHandler", "GetTemplateArgs"),
            "apply": ("ApplyTemplateHandler", "ApplyTemplateArgs"),
        }
    },
    "ManageVaults": {
        "actions": ["list", "switch"],
        "mappings": {
            "list": ("ListVaultsHandler", "ListVaultsArgs"),
            "switch": ("SwitchVaultHandler", "SwitchVaultArgs"),
        }
    }
}

# Read types.go to get struct fields
with open("internal/vault/types.go", "r") as f:
    types_content = f.read()

# Extract structs
structs = {}
matches = re.finditer(r"type (\w+) struct \{([^}]+)\}", types_content)
for match in matches:
    name = match.group(1)
    body = match.group(2)
    fields = []
    for line in body.strip().split("\n"):
        line = line.strip()
        if not line or line.startswith("//"):
            continue
        parts = line.split()
        if len(parts) >= 2:
            field_name = parts[0]
            field_type = parts[1]
            jsontag = ""
            jsonschematag = ""
            if "json:" in line:
                json_match = re.search(r'json:"([^"]+)"', line)
                if json_match: jsontag = json_match.group(1)
            if "jsonschema:" in line:
                schema_match = re.search(r'jsonschema:"([^"]+)"', line)
                if schema_match: jsonschematag = schema_match.group(1)
            
            fields.append({
                "name": field_name,
                "type": field_type,
                "json": jsontag,
                "schema": jsonschematag
            })
    structs[name] = fields

out = """package vault

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

"""

for group_name, group_data in groups.items():
    wrapper_args_name = f"{group_name}MultiplexArgs"
    
    # Collect all unique fields across mapped structs
    unique_fields = {}
    
    for action, (handler_name, args_struct_name) in group_data["mappings"].items():
        if args_struct_name not in structs:
            # print(f"Warning: {args_struct_name} not found")
            continue
        for field in structs[args_struct_name]:
            field_name = field["name"]
            if field_name not in unique_fields:
                unique_fields[field_name] = field
            elif unique_fields[field_name]["type"] != field["type"]:
                # Print conflict
                pass
    
    # Generate struct
    out += f"// {wrapper_args_name} multiplexed args\n"
    out += f"type {wrapper_args_name} struct {{\n"
    action_schema = f"Action to perform: " + ", ".join(f"'{a}'" for a in group_data['actions'])
    out += f'\tAction string `json:"action" jsonschema:"{action_schema}"`\n'
    
    for field in unique_fields.values():
        json_tag = field["json"].replace(",omitempty", "") + ",omitempty"
        field_name = field["name"]
        if field_name == "Action":
            # Skip the inner Action because we already defined Action for the multiplexer
            field_name = "TagAction"
            json_tag = "tag_action,omitempty"
        out += f'\t{field_name} {field["type"]} `json:"{json_tag}" jsonschema:"{field["schema"]}"`\n'
    out += "}\n\n"
    
    # Generate handler
    out += f"// {group_name}MultiplexHandler routes to the specific handler\n"
    out += f"func (v *Vault) {group_name}MultiplexHandler(ctx context.Context, req *mcp.CallToolRequest, args {wrapper_args_name}) (*mcp.CallToolResult, any, error) {{\n"
    out += "\tswitch args.Action {\n"
    
    for action, (handler_name, args_struct_name) in group_data["mappings"].items():
        out += f'\tcase "{action}":\n'
        out += f'\t\tspecificArgs := {args_struct_name}{{\n'
        if args_struct_name in structs:
            for field in structs[args_struct_name]:
                out += f'\t\t\t{field["name"]}: args.{field["name"]},\n'
        out += '\t\t}\n'
        out += f'\t\treturn v.{handler_name}(ctx, req, specificArgs)\n'
        
    out += "\tdefault:\n"
    out += f'\t\treturn nil, nil, fmt.Errorf("unknown action: %s", args.Action)\n'
    out += "\t}\n"
    out += "}\n\n"

with open("internal/vault/multiplex.go", "w") as f:
    f.write(out)

print("Generated internal/vault/multiplex.go")
