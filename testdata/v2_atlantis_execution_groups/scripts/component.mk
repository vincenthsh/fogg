# Auto-generated by fogg. Do not edit
# Make improvements in fogg, so that everyone can benefit.

SELF_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
CHECK_PLANFILE_PATH ?= check-plan.output

include $(SELF_DIR)/common.mk

all:
.PHONY: all

setup: ## set up local dependencies for this repo
	$(MAKE) -C $(REPO_ROOT) setup
.PHONY: setup

check: lint check-plan ## run all checks for this component
.PHONY: check

fmt: terraform ## format code in this component
	$(terraform_command) fmt $(TF_ARGS)
.PHONY: fmt

lint: lint-terraform-fmt lint-tflint ## run all linters for this component
.PHONY: lint

lint-tflint: ## run the tflint linter for this component
	@printf "tflint: "
ifeq ($(TFLINT_ENABLED),1)
	@tflint -c $(REPO_ROOT)/.tflint.hcl || exit $$?;
else
	@echo "disabled"
endif
.PHONY: lint-tflint

lint-terraform-fmt: terraform ## run `terraform fmt` in check mode
	$(terraform_command) fmt $(TF_ARGS) --check=true --diff=true
.PHONY: lint-terraform-fmt

check-auth: check-auth-aws check-auth-heroku ## check that authentication is properly set up for this component
.PHONY: check-auth

check-auth-aws:
	@for p in $(AWS_BACKEND_PROFILE) $(AWS_PROVIDER_PROFILE); do \
		aws --profile $$p sts get-caller-identity > /dev/null || (echo "AWS AUTH error. This component is configured to use a profile named '$$p'. Please add one to your ~/.aws/config" && exit -1); \
	done
	@for r in $(AWS_BACKEND_ROLE_ARN) $(AWS_PROVIDER_ROLE_ARN); do \
		aws sts assume-role --role-arn $$r --role-session-name fogg-auth-test > /dev/null || (echo "AWS AUTH error. This component is configured to use a role named '$$r'." && exit -1); \
	done
.PHONY: check-auth-aws

check-auth-heroku:
ifeq ($(HEROKU_PROVIDER),1)
	@echo "Checking heroku auth..."
	@if command heroku >/dev/null; then \
		heroku auth:whoami || timeout 15 heroku auth:login || (echo "Not authenticated to heroku. For SSO accounts, run 'heroku login', for non-sso accounts set HEROKU_EMAIL and HEROKU_API_KEY" && exit -1); \
	else \
		echo "Heroku CLI not installed, can't check auth."; \
	fi
endif
.PHONY: check-auth-heroku

refresh:
	@if [ "$(TF_BACKEND_KIND)" != "remote" ]; then \
		$(terraform_command) refresh $(TF_ARGS); \
		date +%s > .terraform/refreshed_at; \
	else \
		echo "remote backend does not support the refresh command, skipping"; \
	fi
.PHONY: refresh

refresh-cached:
	@last_refresh=`cat .terraform/refreshed_at 2>/dev/null || echo '0'`; \
	current_time=`date +%s`; \
	if (( current_time - last_refresh > 600 )); then \
		echo "It has been awhile since the last refresh. It is time."; \
		$(MAKE) refresh; \
	else \
		echo "Not time to refresh yet."; \
	fi;
.PHONY: refresh-cached

plan: check-auth init fmt refresh-cached ## run a terraform plan
	$(terraform_command) plan $(TF_ARGS) -refresh=$(REFRESH) -input=false
.PHONY: plan

apply: check-auth init refresh ## run a terraform apply
	@$(terraform_command) apply $(TF_ARGS) -refresh=$(REFRESH) -auto-approve=$(AUTO_APPROVE)
.PHONY: apply

docs:
	echo
.PHONY: docs

clean: ## clean modules and plugins for this component
	-rm -rfv .terraform/modules
	-rm -rfv .terraform/plugins
.PHONY: clean

test:
.PHONY: test

init: terraform check-auth ## run terraform init for this component
	@$(terraform_command) init -input=false
.PHONY: init

check-plan: check-auth init refresh-cached ## run a terraform plan and check that it does not fail
	@if [ "$(TF_BACKEND_KIND)" != "remote" ]; then \
		$(terraform_command) plan $(TF_ARGS) -detailed-exitcode -lock=false -refresh=$(REFRESH) -out=$(CHECK_PLANFILE_PATH) ; \
		ERR=$$?; \
		rm $(CHECK_PLANFILE_PATH) 2>/dev/null; \
	else \
		$(terraform_command) plan $(TF_ARGS) -detailed-exitcode -lock=false; \
		ERR=$$?; \
	fi; \
	if [ $$ERR -eq 0 ] ; then \
		echo "Success"; \
	elif [ $$ERR -eq 1 ] ; then \
		echo "Error in plan execution."; \
		exit 1; \
	elif [ $$ERR -eq 2 ] ; then \
		echo "Diff";  \
	fi;
.PHONY: check-plan

run: check-auth ## run an arbitrary terraform command, CMD. ex `make run CMD='show'`
	@$(terraform_command) $(CMD)
.PHONY: run
