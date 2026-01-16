# GitHub Projects Setup Guide

## Overview

This document describes the GitHub Projects setup for DevTools Sync and how to use it for task tracking.

**Project URL**: https://github.com/users/mark-chris/projects/1

## What's Already Configured

### Custom Fields

The following custom fields have been created and are ready to use:

#### 1. Component (Single Select)
Track which part of the system a task belongs to:
- **Agent** - Go CLI agent work
- **Server** - Backend API work
- **Dashboard** - Frontend React work
- **Infrastructure** - CI/CD, Docker, deployment
- **Documentation** - Docs, guides, examples

#### 2. Priority (Single Select)
Task priority levels:
- **Critical** - Blockers, security issues, production bugs
- **High** - Important features, significant bugs
- **Medium** - Standard features and improvements
- **Low** - Nice-to-haves, minor improvements

#### 3. Effort (Single Select)
Estimated effort or size:
- **XS** - < 2 hours
- **S** - Half day
- **M** - 1-2 days
- **L** - 3-5 days
- **XL** - 1+ week

#### 4. Target Version (Text)
Free-form text for version numbers (e.g., "v1.0", "v1.1", "Beta")

### Repository Integration

- ✅ Project is linked to `mark-chris/devtools-sync`
- ✅ Issues from the repository can be added to the project
- ✅ Built-in automation workflows are enabled

### Automation Workflows

The following workflows are automatically enabled:

1. **Item added to project** - New items are tracked automatically
2. **Item closed** - Closed issues update their status
3. **Pull request linked to issue** - PR associations are tracked
4. **Pull request merged** - Merged PRs can trigger status updates
5. **Auto-add sub-issues** - Sub-issues are automatically added
6. **Auto-close issue** - Issues can be auto-closed based on conditions

## Completing the Setup

The gh CLI has limited support for view creation and status customization. Complete the following steps through the GitHub web interface.

### Step 1: Customize Status Field

By default, the project has three statuses: Todo, In Progress, and Done. Let's add Backlog and In Review.

1. Go to https://github.com/users/mark-chris/projects/1
2. Click the **⚙️ Settings** icon (top right)
3. Scroll to **Status** field settings
4. Click **Edit options**
5. Add two new options:
   - **Backlog** (place it first)
   - **In Review** (place it between "In Progress" and "Done")
6. Optionally, rename **Todo** to **To Do** (with a space)
7. Click **Save**

Your Status field should now have:
- 📋 Backlog
- ✅ To Do
- 🚀 In Progress
- 👀 In Review
- ✔️ Done

### Step 2: Create Board by Status View

Create the primary Kanban-style workflow view:

1. In your project, click **+ New view**
2. Name it: **Board by Status**
3. Layout: **Board**
4. Group by: **Status**
5. Click **Save**
6. Configure the view:
   - Click **⋯** menu → **Sort**
   - Sort by: **Priority** (descending), then **Created** (ascending)
7. Set as default view: Click **⋯** → **Set as default view**

### Step 3: Create Component Table View

Create a spreadsheet view grouped by component:

1. Click **+ New view**
2. Name it: **Component Table**
3. Layout: **Table**
4. Click **Save**
5. Configure the view:
   - Click **Group** → Select **Component**
   - Ensure these columns are visible:
     - Title
     - Status
     - Priority
     - Effort
     - Target Version
     - Assignees
   - Click **Sort** → Sort by **Priority** (descending)

### Step 4: Create Priority Matrix View

Create a board view focused on priorities:

1. Click **+ New view**
2. Name it: **Priority Matrix**
3. Layout: **Board**
4. Click **Save**
5. Configure the view:
   - Click **Group** → Select **Priority**
   - Click **Filter** → Add filter: **Status is not Done**
   - This shows only active work organized by priority

### Step 5: Configure Workflow Automation

Set up automated status transitions:

1. Go to project **Settings** (⚙️ icon)
2. Scroll to **Workflows**
3. Configure these workflows:

**When items are added:**
- Click **Item added to project**
- Set: **Status** = **Backlog**
- Enable the workflow

**When PRs are linked:**
- Click **Pull request linked to issue**
- Set: **Status** = **In Review**
- Enable the workflow

**When PRs are merged:**
- Click **Pull request merged**
- Set: **Status** = **Done**
- Enable the workflow

**When items are closed:**
- Click **Item closed**
- Set: **Status** = **Done**
- Enable the workflow

## Using the Project

### Adding Issues to the Project

**Method 1: From an Issue**
1. Open any issue in the repository
2. Click **Projects** in the right sidebar
3. Select **DevTools Sync Development**
4. The issue will be added with Status = Backlog

**Method 2: From the Project Board**
1. Go to the project
2. Click **+ Add item** in any column
3. Search for or create a new issue
4. Set the Component, Priority, and Effort fields

**Method 3: Automatic Addition**
- Configure the "Item added" workflow to automatically add all new issues

### Working with Tasks

#### Starting Work
1. Move item from **Backlog** → **To Do** when prioritized
2. Move to **In Progress** when you start working
3. Assign yourself to the task

#### Code Review
1. Create a PR and link it to the issue (use "Fixes #123" in PR description)
2. Task automatically moves to **In Review**

#### Completing Work
1. Merge the PR
2. Task automatically moves to **Done**
3. Issue closes automatically

### Best Practices

#### Setting Component
- Always set the Component field for new tasks
- Use **Infrastructure** for CI/CD, Docker, and deployment work
- Use **Documentation** for README, guides, and API docs

#### Setting Priority
- **Critical**: Security issues, production blockers, data loss bugs
- **High**: Important features, bugs affecting core functionality
- **Medium**: Standard feature work, non-critical bugs
- **Low**: Nice-to-haves, polish, minor improvements

#### Setting Effort
- **XS**: Quick fixes, typos, simple config changes
- **S**: Small features, straightforward bug fixes
- **M**: Typical feature work, complex bug fixes
- **L**: Large features, significant refactoring
- **XL**: Major features, architectural changes

#### Using Target Version
- Set this for tasks tied to specific releases
- Use semantic versioning: "v1.0", "v1.1", "v2.0"
- Use labels like "Beta", "Alpha" for pre-release work

### Filtering and Searching

#### Filter by Component
```
is:open component:Agent
is:open component:Server
is:open component:Dashboard
```

#### Filter by Priority
```
is:open priority:Critical
is:open priority:High
```

#### Filter by Effort
```
is:open effort:M
is:open effort:L,XL
```

#### Combined Filters
```
is:open component:Server priority:High status:"In Progress"
```

## View Descriptions

### Board by Status (Default View)
**Purpose**: Primary workflow view for day-to-day task management

**Use when**:
- Planning your daily work
- Moving tasks through the workflow
- Getting an overview of work in progress

**Tips**:
- Limit WIP by keeping "In Progress" column small
- Review "In Review" regularly to unblock PRs
- Groom "Backlog" weekly to move items to "To Do"

### Component Table
**Purpose**: See all work organized by system component

**Use when**:
- Planning sprint or milestone work
- Balancing work across components
- Identifying bottlenecks in specific areas

**Tips**:
- Check regularly to ensure work is balanced
- Use for planning: "We need more Dashboard work"
- Good for multi-person teams with component ownership

### Priority Matrix
**Purpose**: Focus on high-priority active work

**Use when**:
- Need to focus on urgent work
- Planning what to tackle next
- Identifying critical blockers

**Tips**:
- Start your day here to find the most important work
- Keep Critical and High columns as empty as possible
- Use for stakeholder updates on high-priority items

## CLI Commands for Quick Actions

### Add an issue to the project
```bash
gh issue create --title "Task title" --body "Description" --project "DevTools Sync Development"
```

### View project items
```bash
gh project item-list 1 --owner mark-chris
```

### Create issue with labels and project
```bash
gh issue create \
  --title "Add authentication to CLI" \
  --body "Implement OAuth flow for agent authentication" \
  --label "enhancement" \
  --project "DevTools Sync Development"
```

### Link existing issue to project
```bash
gh project item-add 1 --owner mark-chris --url https://github.com/mark-chris/devtools-sync/issues/123
```

## Keyboard Shortcuts

When viewing the project:
- `c` - Create new item
- `e` - Edit item
- `n` - Create new view
- `/` - Focus search
- `?` - Show all shortcuts

## Troubleshooting

### Items not appearing in project
- Check that the issue is in the `mark-chris/devtools-sync` repository
- Verify the project link in the issue sidebar
- Try manually adding with `gh project item-add`

### Automation not working
- Check workflow settings are enabled
- Verify conditions match your workflow
- Check project Settings → Workflows

### Fields not showing in views
- Edit the view and ensure columns are visible
- Check field configuration in Settings
- Refresh the page

## Project Maintenance

### Weekly Review
- Archive or close completed items older than 2 weeks
- Groom Backlog: prioritize or close stale items
- Review In Progress items: identify blockers

### Monthly Review
- Analyze completed work by component
- Adjust priorities based on roadmap
- Update Target Version for upcoming releases

## Resources

- [GitHub Projects Documentation](https://docs.github.com/en/issues/planning-and-tracking-with-projects)
- [GitHub CLI Projects Reference](https://cli.github.com/manual/gh_project)
- [Project URL](https://github.com/users/mark-chris/projects/1)

---

**Last Updated**: 2026-01-16
**Project Created**: 2026-01-16
