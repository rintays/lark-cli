.PHONY: test unit it it-wiki-member-role-update

test: unit

unit:
	go test ./...

# All integration tests (requires explicit opt-in).
it:
	@test "$$LARK_INTEGRATION" = "1" || (echo "LARK_INTEGRATION must be exactly 1 to run integration tests" && exit 1)
	go test ./cmd/lark -run Integration -count=1 -v

# Wiki SpaceMember.Create role-upsert verification (integration).
# Requires:
#   - LARK_INTEGRATION=1
#   - LARK_TEST_WIKI_SPACE_ID
#   - LARK_TEST_USER_EMAIL
it-wiki-member-role-update:
	@test -n "$$LARK_INTEGRATION" || (echo "missing env: LARK_INTEGRATION=1" && exit 1)
	@test -n "$$LARK_TEST_WIKI_SPACE_ID" || (echo "missing env: LARK_TEST_WIKI_SPACE_ID" && exit 1)
	@test -n "$$LARK_TEST_USER_EMAIL" || (echo "missing env: LARK_TEST_USER_EMAIL" && exit 1)
	go test ./cmd/lark -run '^TestWikiMemberRoleUpdateIntegration$$' -count=1 -v
