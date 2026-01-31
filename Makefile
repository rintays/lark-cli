.PHONY: test unit it-wiki-member-role-update

test: unit

unit:
	go test ./...

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
