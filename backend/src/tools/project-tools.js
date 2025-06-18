export const listProjects = {
  name: 'list_projects',
  description: 'List all available Relay projects',
  parameters: {
    type: 'object',
    properties: {
      sort_by: {
        type: 'string',
        enum: ['name', 'last_opened', 'path'],
        description: 'Sort projects by specified field (default: name)'
      },
      include_path: {
        type: 'boolean',
        description: 'Include full file paths in the response'
      }
    }
  },
  async execute(relayManager, projectName, args) {
    try {
      const projects = await relayManager.listProjects();
      
      if (args.sort_by) {
        projects.sort((a, b) => {
          const aVal = a[args.sort_by] || '';
          const bVal = b[args.sort_by] || '';
          return aVal.localeCompare(bVal);
        });
      }
      
      const response = args.include_path ? projects : projects.map(p => ({
        name: p.name,
        lastOpened: p.lastOpened
      }));
      
      return {
        success: true,
        message: `Found ${projects.length} projects`,
        data: { projects: response }
      };
    } catch (error) {
      return {
        success: false,
        message: `Failed to list projects: ${error.message}`
      };
    }
  }
};

export const selectProject = {
  name: 'select_project',
  description: 'Select and switch to a different project',
  parameters: {
    type: 'object',
    properties: {
      name: {
        type: 'string',
        description: 'Name of the project to select'
      }
    },
    required: ['name']
  },
  async execute(relayManager, currentProject, args) {
    try {
      await relayManager.selectProject(args.name);
      return {
        success: true,
        message: `Successfully selected project: ${args.name}`,
        data: { project: args.name }
      };
    } catch (error) {
      return {
        success: false,
        message: `Failed to select project: ${error.message}`
      };
    }
  }
};

export const projectStatus = {
  name: 'project_status',
  description: 'Get detailed status information for the current project',
  parameters: {
    type: 'object',
    properties: {
      include_git: {
        type: 'boolean',
        description: 'Include git status information (default: true)'
      },
      include_issues: {
        type: 'boolean',
        description: 'Include GitHub issues count (default: false)'
      }
    }
  },
  async execute(relayManager, projectName, args) {
    if (!projectName) {
      return {
        success: false,
        message: 'No project currently selected'
      };
    }
    
    try {
      const status = await relayManager.getProjectStatus(projectName);
      
      const response = {
        name: status.name,
        path: status.path
      };
      
      if (args.include_git !== false) {
        response.git = {
          branch: status.git_branch,
          hasChanges: status.has_changes,
          lastCommit: status.last_commit,
          status: status.git_status
        };
      }
      
      if (args.include_issues) {
        const issuesResult = await relayManager.listIssues(projectName);
        if (issuesResult.success) {
          response.issues = {
            count: issuesResult.data.issues.length,
            openCount: issuesResult.data.issues.filter(i => i.state === 'open').length
          };
        }
      }
      
      return {
        success: true,
        message: 'Project status retrieved successfully',
        data: response
      };
    } catch (error) {
      return {
        success: false,
        message: `Failed to get project status: ${error.message}`
      };
    }
  }
};

export const createProject = {
  name: 'create_project',
  description: 'Create a new Relay project',
  parameters: {
    type: 'object',
    properties: {
      name: {
        type: 'string',
        description: 'Name of the new project'
      },
      path: {
        type: 'string',
        description: 'Path where the project should be created'
      },
      template: {
        type: 'string',
        description: 'Project template to use (optional)'
      },
      init_git: {
        type: 'boolean',
        description: 'Initialize as git repository (default: true)'
      }
    },
    required: ['name', 'path']
  },
  async execute(relayManager, currentProject, args) {
    try {
      const result = await relayManager.createProject(args);
      return {
        success: true,
        message: `Project '${args.name}' created successfully`,
        data: result
      };
    } catch (error) {
      return {
        success: false,
        message: `Failed to create project: ${error.message}`
      };
    }
  }
};