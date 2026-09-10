Do not read, grep, cat, or otherwise look at or edit any files or directories that are blocked by `.gitignore`.

NEVER commit code that contains AI agent attribution. Do NOT add the agent name (e.g. Claude, Generated with Claude Code, Co-Authored-By Claude) anywhere in commit messages, PR descriptions, or other Git/GitHub messages.

Git history & merges: Never rewrite or force-push shared history (especially main) unless I ask. Merge PRs with a merge commit -- don't squash, and don't rebase-merge.

Default to using Bun instead of Node.js. That means use commands like `bunx` instead of `npx`.

Write tests for all code.

The task runner is [just](https://github.com/casey/just). `just` script names should be written in `snake_case`.

This site is statically generated so please put trailing slashes on all internal links.

NEVER remove comments from the code without asking. They sometimes contain important notes that are needed for later.

When writing markdown, don't wrap lines in the middle of a paragraph with single line breaks. Let the lines run their full length.

Note that `rm` and `ls` might be overriden with another bash script so you might need to type out the full paths like `/bin/rm` and `/bin/ls` to use the normal commands.

Your Turn Summary: End every substantive reply with a short bulleted list under a bold Your turn heading, covering only what I need to do. It goes last, after everything else in the message. Phrase each bullet as an action I take, and put any link or command I need inside the bullet. Leave out what you already did unless I have to check it. When there is nothing for me to do, say that in one bullet, such as waiting on a check to finish, so a missing list never has to be interpreted. Skip the list only on one-line conversational answers. Keep it to about five bullets. If it runs longer, the message is doing too much.
