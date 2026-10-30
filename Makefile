# qubership-mongodb-operator overrides for the shared consumer Makefile in
# Makefile.precheck. Set values here, then include that file. Do not edit
# Makefile.precheck — refresh it from platform-ci-toolkit/linter/templates/Makefile
# when the template changes:
#   make sync-precheck-makefile PLATFORM_CI_TOOLKIT_RAW_URL=<raw-url>

LINTER_TAG ?= common_linter_integration
DEFAULT_BRANCH ?= main

include $(dir $(abspath $(lastword $(MAKEFILE_LIST))))Makefile.precheck
