# qubership-mongodb-operator overrides for the shared consumer Makefile in
# Makefile.linter. Set values here, then include that file. Makefile.linter sets
# REPO_ROOT from its own path, so this file restores REPO_ROOT afterward.
#
# Refresh Makefile.linter from platform-ci-toolkit when the template changes:
#   make sync-linter-makefile PLATFORM_CI_TOOLKIT_RAW_URL=<raw-url>

REPO_ROOT := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

LINTER_TAG ?= common_linter_integration
DEFAULT_BRANCH ?= main

# Evaluate against this checkout before Makefile.linter overwrites REPO_ROOT.
GO_TEST_MODULES ?= $(shell find $(REPO_ROOT) -name go.mod -not -path '*/vendor/*' -exec dirname {} \;)
JAVA_TEST_MODULES ?= $(shell find $(REPO_ROOT) -name pom.xml -not -path '*/target/*' -exec dirname {} \;)

include $(REPO_ROOT)Makefile.linter

# Point REPO_ROOT back at this repository (firstword is this Makefile).
REPO_ROOT := $(dir $(abspath $(firstword $(MAKEFILE_LIST))))
