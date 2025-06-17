export const createGitHubIssue = {
  name: 'create_github_issue',
  description: 'Create a new GitHub issue for the current project',
  parameters: {
    type: 'object',
    properties: {
      title: {
        type: 'string',
        description: 'The title of the issue'
      },
      body: {
        type: 'string',
        description: 'The body/description of the issue'
      },
      labels: {
        type: 'array',
        items: { type: 'string' },
        description: 'Labels to add to the issue (optional)'
      },
      assignees: {
        type: 'array',
        items: { type: 'string' },
        description: 'GitHub usernames to assign to the issue (optional)'
      },
      priority: {
        type: 'string',
        enum: ['low', 'medium', 'high', 'critical'],
        description: 'Priority level of the issue'
      }
    },
    required: ['title']
  },
  async execute(relayManager, projectName, args) {
    return await relayManager.createGitHubIssue(projectName, args);
  }
};

export const updateGitHubIssue = {
  name: 'update_github_issue',
  description: 'Update an existing GitHub issue',
  parameters: {
    type: 'object',
    properties: {
      number: {
        type: 'number',
        description: 'The issue number to update'
      },
      title: {
        type: 'string',
        description: 'New title for the issue (optional)'
      },
      body: {
        type: 'string',
        description: 'New body for the issue (optional)'
      },
      state: {
        type: 'string',
        enum: ['open', 'closed'],
        description: 'New state for the issue (optional)'
      },
      labels: {
        type: 'array',
        items: { type: 'string' },
        description: 'New labels for the issue (optional)'
      }
    },
    required: ['number']
  },
  async execute(relayManager, projectName, args) {
    return await relayManager.updateGitHubIssue(projectName, args);
  }
};

export const listGitHubIssues = {
  name: 'list_issues',
  description: 'List GitHub issues for the current project',
  parameters: {
    type: 'object',
    properties: {
      state: {
        type: 'string',
        enum: ['open', 'closed', 'all'],
        description: 'Filter issues by state (default: open)'
      },
      assignee: {
        type: 'string',
        description: 'Filter issues by assignee username'
      },
      labels: {
        type: 'array',
        items: { type: 'string' },
        description: 'Filter issues by labels'
      },
      limit: {
        type: 'number',
        description: 'Maximum number of issues to return (default: 10)'
      }
    }
  },
  async execute(relayManager, projectName, args) {
    return await relayManager.listIssues(projectName, args);
  }
};

export const closeGitHubIssue = {
  name: 'close_github_issue',
  description: 'Close a GitHub issue',
  parameters: {
    type: 'object',
    properties: {
      number: {
        type: 'number',
        description: 'The issue number to close'
      },
      reason: {
        type: 'string',
        description: 'Reason for closing the issue (optional)'
      }
    },
    required: ['number']
  },
  async execute(relayManager, projectName, args) {
    return await relayManager.updateGitHubIssue(projectName, {
      ...args,
      state: 'closed'
    });
  }
};