#!/usr/bin/env bash
set -euo pipefail

# Assumes the script is run from the repo root.
output_path="communications/context.txt"

mkdir -p "$(dirname "$output_path")"

prd_content="$(cat specs/PRD.md)"
additional_doc_content="$(cat "specs/Money Coach AI 1.0完整策略库技术实现方案.md")"
prototypes_content="$(cat prototypes/money-coach-v1-prototypes.md)"
backend_spec_content="$(cat prototypes/money-coach-v1-backend-spec.md)"
ci_content="$(cat .github/workflows/ci.yml)"
cloudformation_content="$(cat app-backend/cloudformation/backend.yaml)"

cat > "$output_path" <<EOF
I have the following PRD:

"""
$prd_content
"""

And an additional doc to the PRD:

"""
$additional_doc_content
"""

I wrote the following prototypes docs (mobile app & backend) according to the PRD (and the additional doc):

"""
$prototypes_content
"""

"""
$backend_spec_content
"""

And I've implemented the following CI files to indicate the system architecture and env vars intended to supply:

.github/workflows/ci.yml
"""
$ci_content
"""

app-backend/cloudformation/backend.yaml
"""
$cloudformation_content
"""

I want you to review the prototypes doc and check if it follows the PRD and if there's inconsistency, ambiguity, or under-specifications. The line "Excluded in MVP" is intentional since the lack of corresponding data sources. And check if the CI files are consistent with the docs, and if the env vars are sufficient; and check if every env var is required (optional ones tend to confuse people).
EOF
