"""Ensure tests in this worktree import from the local src/, not any installed package.

When the project is installed as editable (pip install -e .) pointing at the
master worktree, Python resolves `armies` to the master copy rather than this
worktree's copy.  Inserting our own src/ first on sys.path fixes that.
"""

from __future__ import annotations

import sys
from pathlib import Path

# Insert this worktree's src directory at the front of sys.path so that
# `import armies` always resolves to the local source tree.
_src = Path(__file__).parent.parent / "src"
if str(_src) not in sys.path:
    sys.path.insert(0, str(_src))
