import re

with open("internal/vault/multiplex.go", "r") as f:
    text = f.read()

# Rename Args to MultiplexArgs in struct definitions and handler signatures
text = re.sub(r'type (\w+)Args struct', r'type \1MultiplexArgs struct', text)
text = re.sub(r'(args )(\w+)Args', r'\1\2MultiplexArgs', text)

# Fix duplicate Action in BulkOperationsMultiplexArgs
# It currently has:
# Action string `json:"action" jsonschema:"Action to perform: 'tag', 'move', 'set-frontmatter'"`
# Action string `json:"action,omitempty" jsonschema:"Action: 'add' (default), 'remove'"`
bulk_search = r'(Action string `json:"action,omitempty" jsonschema:"Action: \'add\' \(default\), \'remove\'"`)'
text = re.sub(bulk_search, r'TagAction string `json:"tag_action,omitempty" jsonschema:"Tag Action: \'add\' (default), \'remove\'"`', text)

# We also need to fix mapping inside BulkOperationsMultiplexHandler
# Action: args.Action -> Action: args.TagAction
text = re.sub(r'Action: args\.Action,\n(\s+DryRun)', r'Action: args.TagAction,\n\1', text)

# Remove "encoding/json"
text = text.replace('\n    "encoding/json"\n', '')

with open("internal/vault/multiplex.go", "w") as f:
    f.write(text)
