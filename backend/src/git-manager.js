import { exec } from 'child_process';
import { promisify } from 'util';
import fs from 'fs/promises';
import path from 'path';

const execAsync = promisify(exec);

export class GitManager {
  constructor() {
    this.githubToken = process.env.GH_TOKEN;
    this.baseDirectory = process.env.CODE_BASE_DIR || '/home/relay/projects';
    this.gitUsername = process.env.GIT_USERNAME || '';
  }

  async getConfig() {
    return {
      baseDirectory: this.baseDirectory,
      gitUsername: this.gitUsername,
      hasToken: !!this.githubToken
    };
  }

  async setConfig(config) {
    // In a production app, you'd want to persist this to a config file or database
    // For now, we'll just update the in-memory values
    if (config.baseDirectory) {
      this.baseDirectory = config.baseDirectory;
    }
    if (config.gitUsername) {
      this.gitUsername = config.gitUsername;
    }
    
    return await this.getConfig();
  }

  async checkRepositoryStatus(repositories) {
    const statusPromises = repositories.map(async (repo) => {
      const localPath = path.join(this.baseDirectory, repo.name);
      try {
        await fs.access(localPath);
        // Check if it's a git repository
        const gitPath = path.join(localPath, '.git');
        await fs.access(gitPath);
        return {
          ...repo,
          isCloned: true,
          localPath
        };
      } catch {
        return {
          ...repo,
          isCloned: false,
          localPath
        };
      }
    });

    return Promise.all(statusPromises);
  }

  async cloneRepository(repository) {
    if (!this.githubToken) {
      throw new Error('GitHub token required for cloning repositories');
    }

    if (!this.gitUsername) {
      throw new Error('Git username required for cloning repositories');
    }

    const localPath = path.join(this.baseDirectory, repository.name);
    
    try {
      // Ensure base directory exists
      await fs.mkdir(this.baseDirectory, { recursive: true });

      // Check if repository already exists
      try {
        await fs.access(localPath);
        throw new Error('Repository already exists locally');
      } catch (error) {
        if (error.message === 'Repository already exists locally') {
          throw error;
        }
        // Directory doesn't exist, which is what we want
      }

      // Construct clone URL with authentication
      const cloneUrl = `https://${this.gitUsername}:${this.githubToken}@github.com/${repository.fullName}.git`;
      
      console.log(`Cloning repository ${repository.fullName} to ${localPath}`);
      
      const { stdout, stderr } = await execAsync(`git clone "${cloneUrl}" "${localPath}"`, {
        timeout: 30000 // 30 second timeout
      });

      console.log('Clone completed successfully');
      console.log('stdout:', stdout);
      if (stderr) console.log('stderr:', stderr);

      return {
        success: true,
        message: `Repository ${repository.name} cloned successfully`,
        localPath
      };
    } catch (error) {
      console.error('Failed to clone repository:', error);
      
      // Clean up partial clone if it exists
      try {
        await fs.rmdir(localPath, { recursive: true });
      } catch (cleanupError) {
        console.error('Failed to clean up partial clone:', cleanupError);
      }

      return {
        success: false,
        message: `Failed to clone repository: ${error.message}`
      };
    }
  }

  async pullRepository(repository) {
    const localPath = path.join(this.baseDirectory, repository.name);
    
    try {
      // Check if repository exists locally
      await fs.access(localPath);
      const gitPath = path.join(localPath, '.git');
      await fs.access(gitPath);

      console.log(`Pulling latest changes for ${repository.name}`);
      
      const { stdout, stderr } = await execAsync('git pull', {
        cwd: localPath,
        timeout: 30000
      });

      console.log('Pull completed successfully');
      console.log('stdout:', stdout);
      if (stderr) console.log('stderr:', stderr);

      return {
        success: true,
        message: `Repository ${repository.name} updated successfully`,
        output: stdout
      };
    } catch (error) {
      console.error('Failed to pull repository:', error);
      return {
        success: false,
        message: `Failed to update repository: ${error.message}`
      };
    }
  }

  async removeRepository(repositoryName) {
    const localPath = path.join(this.baseDirectory, repositoryName);
    
    try {
      await fs.access(localPath);
      await fs.rmdir(localPath, { recursive: true });

      return {
        success: true,
        message: `Repository ${repositoryName} removed successfully`
      };
    } catch (error) {
      console.error('Failed to remove repository:', error);
      return {
        success: false,
        message: `Failed to remove repository: ${error.message}`
      };
    }
  }

  async getRepositoryInfo(repositoryName) {
    const localPath = path.join(this.baseDirectory, repositoryName);
    
    try {
      // Check if repository exists locally
      await fs.access(localPath);
      const gitPath = path.join(localPath, '.git');
      await fs.access(gitPath);

      // Get git status
      const { stdout: status } = await execAsync('git status --porcelain', {
        cwd: localPath
      });

      // Get current branch
      const { stdout: branch } = await execAsync('git branch --show-current', {
        cwd: localPath
      });

      // Get last commit
      const { stdout: lastCommit } = await execAsync('git log -1 --pretty=format:"%h %s %an %ad" --date=relative', {
        cwd: localPath
      });

      return {
        success: true,
        data: {
          localPath,
          isClean: status.trim() === '',
          currentBranch: branch.trim(),
          lastCommit: lastCommit.trim(),
          hasUncommittedChanges: status.trim() !== ''
        }
      };
    } catch (error) {
      return {
        success: false,
        message: `Repository not found locally or not a git repository: ${error.message}`
      };
    }
  }
}