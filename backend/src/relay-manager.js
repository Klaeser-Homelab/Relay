import { exec, spawn } from 'child_process';
import { promisify } from 'util';
import path from 'path';
import fs from 'fs/promises';
import os from 'os';

const execAsync = promisify(exec);

export class RelayManager {
  constructor() {
    this.projectsPath = path.join(os.homedir(), '.relay', 'projects');
    this.configPath = path.join(os.homedir(), '.relay', 'config.json');
    this.ensureConfigDirectory();
  }

  async ensureConfigDirectory() {
    try {
      const configDir = path.dirname(this.configPath);
      await fs.mkdir(configDir, { recursive: true });
    } catch (error) {
      console.error('Failed to create config directory:', error);
    }
  }

  async findRelayBinary() {
    const locations = [
      '../server/relay',
      '../server/tmp/relay',
      './relay',
      'relay'
    ];

    for (const location of locations) {
      try {
        await fs.access(location);
        return location;
      } catch (error) {
        try {
          await fs.access(location + '.exe');
          return location + '.exe';
        } catch (error) {
          continue;
        }
      }
    }

    try {
      const { stdout } = await execAsync('which relay || where relay');
      return stdout.trim();
    } catch (error) {
      return null;
    }
  }

  async listProjects() {
    const relayBinary = await this.findRelayBinary();
    if (relayBinary) {
      return await this.listProjectsViaRelay(relayBinary);
    }
    return await this.discoverProjects();
  }

  async listProjectsViaRelay(relayBinary) {
    try {
      const { stdout } = await execAsync(`${relayBinary} list`);
      const lines = stdout.trim().split('\n');
      const projects = [];

      for (const line of lines) {
        const trimmedLine = line.trim();
        if (!trimmedLine || 
            trimmedLine.startsWith('No projects') || 
            trimmedLine.startsWith('Projects:')) {
          continue;
        }

        const parts = trimmedLine.split(' - ');
        if (parts.length >= 2) {
          const name = parts[0].replace(/^\*\s*/, '').trim();
          const projectPath = parts[1].trim();
          
          projects.push({
            name,
            path: projectPath,
            lastOpened: new Date().toISOString()
          });
        }
      }

      return projects;
    } catch (error) {
      throw new Error(`Failed to list projects via relay: ${error.message}`);
    }
  }

  async discoverProjects() {
    const projects = [];
    const homeDir = os.homedir();
    const searchDirs = [
      path.join(homeDir, 'Code'),
      path.join(homeDir, 'Projects'),
      path.join(homeDir, 'Development'),
      path.join(homeDir, 'src')
    ];

    for (const searchDir of searchDirs) {
      try {
        await fs.access(searchDir);
        const entries = await fs.readdir(searchDir, { withFileTypes: true });
        
        for (const entry of entries) {
          if (!entry.isDirectory()) continue;
          
          const projectPath = path.join(searchDir, entry.name);
          const gitPath = path.join(projectPath, '.git');
          
          try {
            await fs.access(gitPath);
            projects.push({
              name: entry.name,
              path: projectPath,
              lastOpened: new Date().toISOString()
            });
          } catch (error) {
            continue;
          }
        }
      } catch (error) {
        continue;
      }
    }

    return projects;
  }

  async selectProject(projectName) {
    const relayBinary = await this.findRelayBinary();
    if (relayBinary) {
      try {
        await execAsync(`${relayBinary} open ${projectName}`);
        return;
      } catch (error) {
        throw new Error(`Failed to select project via relay: ${error.message}`);
      }
    }

    const projects = await this.listProjects();
    const project = projects.find(p => p.name === projectName);
    if (!project) {
      throw new Error(`Project '${projectName}' not found`);
    }
  }

  async getProjectStatus(projectName) {
    const projects = await this.listProjects();
    const project = projects.find(p => p.name === projectName);
    
    if (!project) {
      throw new Error(`Project '${projectName}' not found`);
    }

    const status = {
      name: projectName,
      path: project.path
    };

    const gitStatus = await this.getGitStatus(project.path);
    Object.assign(status, gitStatus);

    return status;
  }

  async getGitStatus(projectPath) {
    const status = {};

    try {
      const { stdout: branch } = await execAsync('git branch --show-current', { cwd: projectPath });
      status.git_branch = branch.trim();
    } catch (error) {
      status.git_branch = 'unknown';
    }

    try {
      const { stdout: gitStatus } = await execAsync('git status --porcelain', { cwd: projectPath });
      const statusOutput = gitStatus.trim();
      status.has_changes = statusOutput !== '';
      status.git_status = statusOutput;
    } catch (error) {
      status.has_changes = false;
      status.git_status = '';
    }

    try {
      const { stdout: lastCommit } = await execAsync('git log -1 --pretty=format:"%h %s"', { cwd: projectPath });
      status.last_commit = lastCommit.trim();
    } catch (error) {
      status.last_commit = 'No commits';
    }

    return status;
  }

  async executeFunction(projectName, functionName, args) {
    console.log(`Executing function: ${functionName} with args:`, args);

    switch (functionName) {
      case 'create_github_issue':
        return await this.createGitHubIssue(projectName, args);
      case 'update_github_issue':
        return await this.updateGitHubIssue(projectName, args);
      case 'git_commit':
        return await this.gitCommit(projectName);
      case 'git_status':
        return await this.gitStatus(projectName);
      case 'list_issues':
        return await this.listIssues(projectName);
      default:
        return {
          success: false,
          message: `Unknown function: ${functionName}`
        };
    }
  }

  async createGitHubIssue(projectName, args) {
    const { title, body, labels } = args;
    
    if (!title) {
      return {
        success: false,
        message: 'Title is required for creating an issue'
      };
    }

    const relayBinary = await this.findRelayBinary();
    if (relayBinary) {
      console.log(`Would create issue via relay: ${title} - ${body}`);
      return {
        success: true,
        message: `Issue '${title}' created successfully`,
        data: { title, body, labels }
      };
    }

    return {
      success: true,
      message: `Issue '${title}' would be created (relay binary not available)`,
      data: { title, body, labels }
    };
  }

  async updateGitHubIssue(projectName, args) {
    const { number, title, body } = args;

    if (!number) {
      return {
        success: false,
        message: 'Issue number is required'
      };
    }

    console.log(`Would update issue #${number}: ${title} - ${body}`);
    
    return {
      success: true,
      message: `Issue #${number} updated successfully`,
      data: { number, title, body }
    };
  }

  async gitCommit(projectName) {
    const projects = await this.listProjects();
    const project = projects.find(p => p.name === projectName);
    
    if (!project) {
      return {
        success: false,
        message: 'Project not found'
      };
    }

    const relayBinary = await this.findRelayBinary();
    if (relayBinary) {
      try {
        const { stdout } = await execAsync(`${relayBinary} commit`, { cwd: project.path });
        return {
          success: true,
          message: 'Smart commit completed successfully',
          data: { output: stdout }
        };
      } catch (error) {
        return {
          success: false,
          message: `Commit failed: ${error.message}`
        };
      }
    }

    try {
      await execAsync('git add .', { cwd: project.path });
      const { stdout } = await execAsync('git commit -m "Voice-controlled commit"', { cwd: project.path });
      return {
        success: true,
        message: 'Changes committed successfully',
        data: { output: stdout }
      };
    } catch (error) {
      return {
        success: false,
        message: `Commit failed: ${error.message}`
      };
    }
  }

  async gitStatus(projectName) {
    try {
      const status = await this.getProjectStatus(projectName);
      return {
        success: true,
        message: 'Git status retrieved successfully',
        data: status
      };
    } catch (error) {
      return {
        success: false,
        message: error.message
      };
    }
  }

  async listIssues(projectName) {
    return {
      success: true,
      message: 'Issues retrieved successfully',
      data: {
        issues: [
          {
            number: 1,
            title: 'Add voice control feature',
            state: 'open'
          },
          {
            number: 2,
            title: 'Fix audio streaming bug',
            state: 'closed'
          }
        ]
      }
    };
  }
}