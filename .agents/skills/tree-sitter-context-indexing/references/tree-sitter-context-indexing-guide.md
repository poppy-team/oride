# Tree-Sitter Context Indexing — Technical Reference Guide

## 1. Core Concepts

### 1.1 Grammars and Concrete Syntax Trees
tree-sitter compiles one grammar per language into an incremental parser that produces a concrete syntax tree with every token present, including punctuation. Because the tree is concrete, byte offsets map exactly back to source ranges, which is what makes syntax-aligned chunking trustworthy.

### 1.2 Declarative Queries with Predicates
Extraction is declared in `.scm` query files, not imperative tree walks. A Python definition query looks like:

```scm
(function_definition
  name: (identifier) @function.def)
(class_definition
  name: (identifier) @class.def)
(import_statement
  name: (dotted_name) @import.ref)
((identifier) @call.ref
 (#match? @call.ref "^route_"))
```

The `#match?` predicate narrows captures with a regex, while `#eq?` pins exact text. Queries are versioned beside the spec so grammar bumps and query edits are reviewed together.

### 1.3 Incremental Parsing with Old Trees
Re-parsing after an edit reuses every unchanged subtree when the previous tree and the byte-range edit are supplied:

```python
from tree_sitter import Language, Parser
import tree_sitter_python as tspy

lang = Language(tspy.language())
parser = Parser(lang)
old_tree = parser.parse(bytes(source_v1, "utf8"))
new_tree = parser.parse(bytes(source_v2, "utf8"), old_tree=old_tree)
```

Passing `old_tree` turns a 2000-line re-parse into a millisecond patch application. Omitting it silently pays full parse cost on every keystroke.

### 1.4 Chunking by Node Ranges
Each top-level definition node becomes one chunk keyed by its start and end byte offsets. Bodies over the byte budget split at nested definition boundaries, and the enclosing signature line travels as the chunk header so retrieved chunks stay self-describing inside the agent window.

## 2. CLI Workflow
- `tree-sitter parse --quiet file.py` validates grammar quality file by file.
- `tree-sitter query defs.scm file.py` tests captures before they enter the index.
- `tree-sitter highlight --scope source.python file.py` smoke-tests node names after grammar bumps.

## 3. Common Pitfalls
- Assuming node names are stable across grammar versions; a minor bump can rename `dotted_name` and silently empty the import capture.
- Chunking by line windows instead of node ranges, which slices decorators from functions and conditions from branches.
- Treating tree-sitter output as typed semantics; it cannot resolve overloads, imports, or dynamic dispatch.
