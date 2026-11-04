# qubership-mongodb-operator overrides for the shared consumer Makefile in
# common.mk. Set values here, then include that file. Do not edit
# common.mk — refresh it from platform-ci-toolkit/linter/templates/common.mk
# when the template changes.

LINTER_TAG ?= common_linter_integration
DEFAULT_BRANCH ?= main

include common.mk
