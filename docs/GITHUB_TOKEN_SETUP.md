# GitHub token setup for a new team member

Each team member must create their own GitHub personal access token (PAT). Do
not create or share a token from another person's account: a PAT acts with the
permissions of the account that created it.

This project uses the token to identify the user through GitHub and, for some
workflows, to read or update the repository. Prefer a fine-grained PAT because
it can be limited to this organization, this repository, and only the required
permissions.

## Before creating the token

An organization or repository administrator must:

1. Invite the member's GitHub account to the `Euclides-EM` organization and
   give it access to the `commentaria-hub` repository.
2. Add the member's GitHub login, in lowercase, to `allowList` in
   `ocrflow/pkg/httpwrapper/auth.go`, then deploy the backend change.

The second step is required for the Commentaria Hub login. A valid token from a
user who is not in that allowlist is rejected.

## Create a fine-grained PAT

The new member must perform these steps while signed in to their own GitHub
account:

1. Open [GitHub fine-grained token settings](https://github.com/settings/personal-access-tokens/new).
2. Enter a descriptive name, such as `commentaria-hub-laptop`.
3. Choose a short expiration period. Ninety days is a reasonable default unless
   the organization has a different policy.
4. Set **Resource owner** to `Euclides-EM`.
5. Under **Repository access**, choose **Only select repositories**, then select
   `commentaria-hub`.
6. Under **Repository permissions**, grant only what the member needs:

   | Use | Permissions |
   | --- | --- |
   | Sign in to Commentaria Hub and use protected Hub API operations | **Contents: Read-only** |
   | Create branches, push code, and create pull requests from the application | **Contents: Read and write** and **Pull requests: Read and write** |
   | Edit files under `.github/workflows/` | Add **Workflows: Read and write** |

   GitHub automatically includes read-only metadata access.
7. Click **Generate token**.
8. If the organization requires approval, wait until an organization owner
   approves the request. A pending token cannot access private organization
   resources.
9. Copy the token immediately. GitHub displays it only once.

GitHub's current UI and permission descriptions are documented in
[Managing your personal access tokens](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens).

## Store and use the token

Treat the token like a password:

- Never commit it to Git, put it in documentation, or send it in chat or email.
- Paste it into the Commentaria Hub login form when prompted.
- For local development, store it as `GITHUB_TOKEN` in the untracked
  `.env_private` file used by this project.
- Use the operating system's credential manager for Git over HTTPS when
  possible. GitHub recommends GitHub CLI or Git Credential Manager instead of
  manually storing a PAT for command-line Git.

To verify the token without putting its value in shell history:

```bash
read -s GITHUB_TOKEN
echo
curl --fail-with-body \
  -H "Authorization: Bearer $GITHUB_TOKEN" \
  -H "Accept: application/vnd.github+json" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  https://api.github.com/user
unset GITHUB_TOKEN
```

The response should contain the member's GitHub `login`. The Commentaria Hub
login will still fail until that login is present in the backend allowlist.

## Organization approval and SSO

- A fine-grained PAT may remain **Pending** until an organization owner approves
  it in the organization's token settings.
- If the organization uses SAML single sign-on, follow GitHub's authorization
  prompt. Classic PATs require the member to select **Configure SSO** after the
  token is created.
- Organization owners can restrict PAT types, maximum lifetimes, and required
  approvals. Follow the organization policy if it is stricter than this guide.

## When a classic PAT is necessary

Use a classic PAT only when fine-grained PATs cannot support the required
operation. For repository command-line access, the classic token normally needs
the broad `repo` scope. Organization policy may prohibit classic PATs entirely.

Fine-grained PATs currently have limitations for outside collaborators and some
GitHub features. See GitHub's
[fine-grained PAT limitations](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens#fine-grained-personal-access-tokens-limitations)
before falling back to a classic token.

## Rotation and offboarding

- Replace the token before it expires and update every local place where it is
  stored.
- Revoke a token immediately if it may have leaked.
- When a member leaves the team, remove their GitHub organization/repository
  access and remove their login from the backend allowlist. The member should
  also delete the token from their
  [personal access token settings](https://github.com/settings/tokens).
