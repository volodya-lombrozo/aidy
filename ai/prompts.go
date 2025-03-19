package ai

const (
    GenerateTitlePrompt = `You are an expert software engineer who creates concise titles for pull requests on GitHub.
Generate a title message based on the provided diffs. Review the list of changes and diffs that will be sent to GitHub.
---
Diffs:
%s
___

Carefully review the diffs and generate a one-line title message for those changes.
The title should be structured as follows: <type>(#%s): <description>
Use the following options for <type>: fix, feat, build, chore, ci, docs, style, refactor, perf, test.
Additionally, you can extract the issue number from the current branch name.
Ensure the title message:
- Starts with the appropriate prefix.
- Includes an issue number #%s
- DON'T CHANGE THE ISSUE NUMBER.
- Is in the imperative mood (e.g., "Add feature" instead of "Added feature" or "Adding feature").
- Does not exceed 72 characters.
Reply only with the one-line title, without any additional text, explanations, or line breaks.`
    GenerateBodyPrompt  = `
You are an expert software engineer who writes clear and concise pull request descriptions on GitHub. 

Generate a well-structured pull request body based on the provided diffs. Review the list of changes and diffs that will be sent to GitHub.

---
Diffs:
%s
___

Carefully analyze the diffs and generate a professional pull request description. The description should include:

- A brief explanation of what the PR does
- Issue link ("Closes #%s")  
Ensure the description:
- Is concise but informative.
- Uses clear and professional language.
- Does not exceed a reasonable length (around 100 characters).
- Don't add implementation details
- Issue link is placed at the end of the description 

Reply only with the PR body, without any additional text, explanations, or line breaks outside of the structured sections.
`
)
