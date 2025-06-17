export const gitStatus = {
  name: 'git_status',
  description: 'Get the current git status of the project',
  parameters: {
    type: 'object',
    properties: {
      verbose: {
        type: 'boolean',
        description: 'Include detailed file status information'
      }
    }
  },
  async execute(relayManager, projectName, args) {
    return await relayManager.gitStatus(projectName, args);
  }
};

export const gitCommit = {
  name: 'git_commit',
  description: 'Create a smart git commit with automatically generated commit message',
  parameters: {
    type: 'object',
    properties: {
      message: {
        type: 'string',
        description: 'Custom commit message (optional - will auto-generate if not provided)'
      },
      add_all: {
        type: 'boolean',
        description: 'Add all changed files before committing (default: true)'
      },
      dry_run: {
        type: 'boolean',
        description: 'Show what would be committed without actually committing'
      }
    }
  },
  async execute(relayManager, projectName, args) {
    return await relayManager.gitCommit(projectName, args);
  }
};

export const gitPush = {
  name: 'git_push',
  description: 'Push local commits to the remote repository',
  parameters: {
    type: 'object',
    properties: {
      remote: {
        type: 'string',
        description: 'Remote name (default: origin)'
      },
      branch: {
        type: 'string',
        description: 'Branch name (default: current branch)'
      },
      force: {
        type: 'boolean',
        description: 'Force push (use with caution)'
      }
    }
  },
  async execute(relayManager, projectName, args) {
    return await relayManager.gitPush(projectName, args);
  }
};

export const gitPull = {
  name: 'git_pull',
  description: 'Pull latest changes from the remote repository',
  parameters: {
    type: 'object',
    properties: {
      remote: {
        type: 'string',
        description: 'Remote name (default: origin)'
      },
      branch: {
        type: 'string',
        description: 'Branch name (default: current branch)'
      },
      rebase: {
        type: 'boolean',
        description: 'Use rebase instead of merge'
      }
    }
  },
  async execute(relayManager, projectName, args) {
    return await relayManager.gitPull(projectName, args);
  }
};

export const gitBranch = {
  name: 'git_branch',
  description: 'List, create, or switch git branches',
  parameters: {
    type: 'object',
    properties: {
      action: {
        type: 'string',
        enum: ['list', 'create', 'switch', 'delete'],
        description: 'Action to perform with branches'
      },
      name: {
        type: 'string',
        description: 'Branch name (required for create, switch, delete actions)'
      },
      base: {
        type: 'string',
        description: 'Base branch for new branch (default: current branch)'
      }
    },
    required: ['action']
  },
  async execute(relayManager, projectName, args) {
    return await relayManager.gitBranch(projectName, args);
  }
};

export const gitLog = {
  name: 'git_log',
  description: 'Show git commit history',
  parameters: {
    type: 'object',
    properties: {
      limit: {
        type: 'number',
        description: 'Number of commits to show (default: 10)'
      },
      oneline: {
        type: 'boolean',
        description: 'Show compact one-line format'
      },
      author: {
        type: 'string',
        description: 'Filter commits by author'
      },
      since: {
        type: 'string',
        description: 'Show commits since date (e.g., "2 weeks ago")'
      }
    }
  },
  async execute(relayManager, projectName, args) {
    return await relayManager.gitLog(projectName, args);
  }
};