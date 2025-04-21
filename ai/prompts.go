package ai

const (
	GenerateTitlePrompt = `You are an expert software engineer who creates concise and informative pull request titles for GitHub.

Your task is to generate a one-line PR title based on the provided diff and issue description.

<diff>
%s
</diff>

This pull request addresses the following issue:

<issue>
%s
</issue>

Carefully review both the diff and the issue description. Then, generate a PR title in the following format:

<type>(#%s): <description>

- Valid <type> values: fix, feat, build, chore, ci, docs, style, refactor, perf, test.
- Always include the issue number #%s in the title.
- Use the imperative mood (e.g., "add feature", not "added feature" or "adding feature").
- Keep the title within 72 characters.
- Do not include explanations, comments, or line breaks. Return only the title line.

Also, this is the project summary for which you need to create the PR title:
<summary>
%s
</summary>
`

	GenerateBodyPrompt = `You are an expert software engineer who writes clear and professional pull request descriptions on GitHub.

Your task is to generate a well-structured pull request body based on the provided diff and issue description.

<diff>
%s
</diff>

This pull request addresses the following issue:

<issue>
%s
</issue>

Carefully analyze the diff and issue. Then, generate a concise and informative pull request description.

The description must:
- Briefly explain what the PR does (one or two sentences).
- Be clear and professional.
- Avoid listing individual changes.
- Avoid implementation details.
- Avoid explaining why the changes were made.
- Be approximately 50–100 characters long.
- Use backticks for code or identifiers when appropriate.
- End with an empty line followed by: "Closes #%s"

Formatting rules:
- Do not add section headers.
- Do not include line breaks except before the "Closes" line.
- Reply only with the pull request body — no additional text or explanations.

Also, this is the project summary for which you need to create the PR body:
<summary>
%s
</summary>
`

	GenerateCommitPrompt = `You are an expert software engineer who writes concise, one-line Git commit messages based on code diffs.

Your task is to generate a single-line commit message for the following changes. 

<diff>
%s
</diff>

Review diffs carefully.

The commit message must follow this format:
<type>(#%s): <description>

Where:
- <type> is one of: fix, feat, build, chore, ci, docs, style, refactor, perf, test
- #%s is the issue number (do not modify it)

Ensure the message:
- Starts with the appropriate prefix
- Uses the imperative mood (e.g., "Add feature", not "Added feature" or "Adding feature")
- Does not exceed 72 characters

Reply with the commit message only — no explanations, comments, or line breaks.`

	GenerateIssueTitlePrompt = `You are an expert software engineer who writes clear and concise titles for GitHub issues.

Generate a one-line issue title based on the following user input:

<input>
%s
</input>

The title should:
- Clearly reflect the core problem or task
- Be concise and informative
- Use backticks for code identifiers or technical terms, where appropriate
- Not exceed 72 characters

Reply only with the issue title — no explanations, comments, or line breaks.

Also, this is the project summary for which you need to create the issue title:
<summary>
%s
</summary>
`

	GenerateIssueBodyPrompt = `You are an expert software engineer who writes clear and informative descriptions for GitHub issues.

Generate an issue body based on the following user input:

<input>
%s
</input>

The description should:
- Clearly explain the issue
- Be no longer than 200 characters
- Avoid repetition
- Contain no headers
- Be grammatically correct
- Use backticks for code or technical terms where appropriate

Reply only with the issue body — no explanations, comments, or extra formatting.

Also, this is the project summary for which you need to create the issue body:
<summary>
%s
</summary>

`

	GenerateLabelsPrompt = `You are an expert software engineer who understands how to assign appropriate labels to GitHub issues.

Your task is to select the most relevant labels for the following issue:

<issue>
%s
</issue>

Available labels:

<labels>
%s
</labels>

Reply only with the list of selected labels — no explanations, comments, or additional formatting.`

    SummaryPrompt = `Read the following README.md of a software project and generate a short, single-paragraph summary suitable for AI agents.
Focus on key features, purpose, technologies used, and any setup or usage highlights. 
The summary must be concise, comprehensive, and free of any additional commentary or explanation — just the raw summary text.

<readme>
%s
</readme>

Reply only with the issue body — no explanations, comments, or extra formatting.`
)
