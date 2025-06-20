import { Octokit } from '@octokit/rest';
import { GoogleGenerativeAI } from '@google/generative-ai';
import { ClaudeCodeManager } from './claude-code-manager.js';

export class GitHubManager {
  constructor() {
    this.githubToken = process.env.GH_TOKEN;
    this.geminiApiKey = process.env.GEMINI_API_KEY;
    
    if (!this.githubToken) {
      console.warn('GH_TOKEN not provided. GitHub functionality will be limited.');
    }
    
    if (!this.geminiApiKey) {
      console.warn('GEMINI_API_KEY not provided. Implementation advice functionality will be limited.');
    }
    
    this.octokit = new Octokit({
      auth: this.githubToken
    });
    
    this.genAI = this.geminiApiKey ? new GoogleGenerativeAI(this.geminiApiKey) : null;
    this.claudeCodeManager = new ClaudeCodeManager();
    
    this.currentRepository = null;
  }

  async listProjects() {
    if (!this.githubToken) {
      throw new Error('GitHub token required to list repositories');
    }

    try {
      console.log('Fetching GitHub repositories...');
      
      // Get all repositories for the authenticated user
      const { data: repos } = await this.octokit.rest.repos.listForAuthenticatedUser({
        sort: 'updated',
        per_page: 100,
        type: 'all'
      });

      const projects = repos.map(repo => ({
        name: repo.name,
        fullName: repo.full_name,
        path: `github:${repo.full_name}`,
        description: repo.description,
        url: repo.html_url,
        cloneUrl: repo.clone_url,
        defaultBranch: repo.default_branch,
        isPrivate: repo.private,
        language: repo.language,
        lastOpened: repo.updated_at,
        stars: repo.stargazers_count,
        forks: repo.forks_count
      }));

      console.log(`Found ${projects.length} GitHub repositories`);
      return projects;
    } catch (error) {
      console.error('Failed to fetch GitHub repositories:', error);
      throw new Error(`GitHub API error: ${error.message}`);
    }
  }

  async selectProject(projectName) {
    try {
      // projectName can be either "repo-name" or "owner/repo-name"
      let fullName = projectName;
      
      if (!projectName.includes('/')) {
        // Get current user to construct full name
        const { data: user } = await this.octokit.rest.users.getAuthenticated();
        fullName = `${user.login}/${projectName}`;
      }

      // Verify the repository exists and user has access
      const { data: repo } = await this.octokit.rest.repos.get({
        owner: fullName.split('/')[0],
        repo: fullName.split('/')[1]
      });

      this.currentRepository = {
        name: repo.name,
        fullName: repo.full_name,
        owner: repo.owner.login,
        url: repo.html_url,
        defaultBranch: repo.default_branch
      };

      console.log(`Selected repository: ${this.currentRepository.fullName}`);
      return this.currentRepository;
    } catch (error) {
      console.error('Failed to select repository:', error);
      throw new Error(`Repository '${projectName}' not found or access denied`);
    }
  }

  async getProjectStatus(projectName) {
    await this.selectProject(projectName);
    
    if (!this.currentRepository) {
      throw new Error('No repository selected');
    }

    try {
      const [repoInfo, commits, issues] = await Promise.all([
        this.octokit.rest.repos.get({
          owner: this.currentRepository.owner,
          repo: this.currentRepository.name
        }),
        this.octokit.rest.repos.listCommits({
          owner: this.currentRepository.owner,
          repo: this.currentRepository.name,
          per_page: 1
        }),
        this.octokit.rest.issues.listForRepo({
          owner: this.currentRepository.owner,
          repo: this.currentRepository.name,
          state: 'open',
          per_page: 1
        })
      ]);

      const repo = repoInfo.data;
      const latestCommit = commits.data[0];
      const openIssuesCount = issues.data.length;

      return {
        name: repo.name,
        fullName: repo.full_name,
        path: `github:${repo.full_name}`,
        url: repo.html_url,
        description: repo.description,
        defaultBranch: repo.default_branch,
        isPrivate: repo.private,
        language: repo.language,
        stars: repo.stargazers_count,
        forks: repo.forks_count,
        lastCommit: latestCommit ? {
          sha: latestCommit.sha.substring(0, 7),
          message: latestCommit.commit.message.split('\n')[0],
          author: latestCommit.commit.author.name,
          date: latestCommit.commit.author.date
        } : null,
        openIssues: openIssuesCount,
        updatedAt: repo.updated_at
      };
    } catch (error) {
      console.error('Failed to get repository status:', error);
      throw new Error(`Failed to get status for repository: ${error.message}`);
    }
  }

  async executeFunction(projectName, functionName, args, onStreamMessage = null) {
    console.log(`Executing GitHub function: ${functionName} with args:`, args);

    if (projectName) {
      await this.selectProject(projectName);
    }

    switch (functionName) {
      case 'create_github_issue':
        return await this.createGitHubIssue(args);
      case 'update_github_issue':
        return await this.updateGitHubIssue(args);
      case 'list_issues':
        return await this.listIssues(args);
      case 'close_github_issue':
        return await this.closeGitHubIssue(args);
      case 'add_issue_comment':
        return await this.addIssueComment(args);
      case 'get_repository_info':
        return await this.getRepositoryInfo();
      case 'list_commits':
        return await this.listCommits(args);
      case 'create_pull_request':
        return await this.createPullRequest(args);
      case 'list_pull_requests':
        return await this.listPullRequests(args);
      case 'get_implementation_advice':
        return await this.getImplementationAdvice(args);
      case 'ask_claude_to_make_plan':
        return await this.askClaudeToMakePlan(args, onStreamMessage);
      default:
        return {
          success: false,
          message: `Unknown function: ${functionName}`
        };
    }
  }

  async createGitHubIssue(args) {
    if (!this.githubToken) {
      return {
        success: false,
        message: 'GitHub token not configured. Set GH_TOKEN environment variable.'
      };
    }

    if (!this.currentRepository) {
      return {
        success: false,
        message: 'No repository selected'
      };
    }

    const { title, body, labels, assignees } = args;
    
    if (!title) {
      return {
        success: false,
        message: 'Title is required for creating an issue'
      };
    }

    try {
      console.log(`Creating GitHub issue in ${this.currentRepository.fullName}:`, { title, body, labels, assignees });
      
      const { data: issue } = await this.octokit.rest.issues.create({
        owner: this.currentRepository.owner,
        repo: this.currentRepository.name,
        title,
        body: body || '',
        labels: labels || [],
        assignees: assignees || []
      });

      console.log(`Successfully created GitHub issue #${issue.number}:`, issue.html_url);

      return {
        success: true,
        message: `Issue #${issue.number} '${title}' created successfully`,
        data: {
          number: issue.number,
          title: issue.title,
          url: issue.html_url,
          state: issue.state
        }
      };
    } catch (error) {
      console.error('Failed to create GitHub issue:', error);
      console.error('Error details:', error.response?.data || error.message);
      return {
        success: false,
        message: `Failed to create issue: ${error.message}`
      };
    }
  }

  async updateGitHubIssue(args) {
    if (!this.currentRepository) {
      return {
        success: false,
        message: 'No repository selected'
      };
    }

    const { number, title, body, state, labels } = args;

    if (!number) {
      return {
        success: false,
        message: 'Issue number is required'
      };
    }

    try {
      const updateData = {};
      if (title) updateData.title = title;
      if (body) updateData.body = body;
      if (state) updateData.state = state;
      if (labels) updateData.labels = labels;

      const { data: issue } = await this.octokit.rest.issues.update({
        owner: this.currentRepository.owner,
        repo: this.currentRepository.name,
        issue_number: number,
        ...updateData
      });

      return {
        success: true,
        message: `Issue #${issue.number} updated successfully`,
        data: {
          number: issue.number,
          title: issue.title,
          url: issue.html_url,
          state: issue.state
        }
      };
    } catch (error) {
      console.error('Failed to update GitHub issue:', error);
      return {
        success: false,
        message: `Failed to update issue: ${error.message}`
      };
    }
  }

  async closeGitHubIssue(args) {
    return await this.updateGitHubIssue({
      ...args,
      state: 'closed'
    });
  }

  async addIssueComment(args) {
    if (!this.currentRepository) {
      return {
        success: false,
        message: 'No repository selected'
      };
    }

    const { number, body } = args;

    if (!number || !body) {
      return {
        success: false,
        message: 'Issue number and comment body are required'
      };
    }

    try {
      console.log(`Adding comment to issue #${number} in ${this.currentRepository.fullName}`);
      
      const { data: comment } = await this.octokit.rest.issues.createComment({
        owner: this.currentRepository.owner,
        repo: this.currentRepository.name,
        issue_number: number,
        body
      });

      console.log(`Successfully added comment to issue #${number}`);

      return {
        success: true,
        message: `Comment added to issue #${number}`,
        data: {
          id: comment.id,
          url: comment.html_url,
          issueNumber: number
        }
      };
    } catch (error) {
      console.error('Failed to add issue comment:', error);
      return {
        success: false,
        message: `Failed to add comment: ${error.message}`
      };
    }
  }

  async listIssues(args = {}) {
    if (!this.currentRepository) {
      return {
        success: false,
        message: 'No repository selected'
      };
    }

    try {
      const { state = 'open', limit = 10, assignee, labels } = args;
      
      const params = {
        owner: this.currentRepository.owner,
        repo: this.currentRepository.name,
        state,
        per_page: limit
      };

      if (assignee) params.assignee = assignee;
      if (labels && labels.length > 0) params.labels = labels.join(',');

      const { data: issues } = await this.octokit.rest.issues.listForRepo(params);

      const issueList = issues.map(issue => ({
        number: issue.number,
        title: issue.title,
        state: issue.state,
        url: issue.html_url,
        author: issue.user.login,
        labels: issue.labels.map(label => label.name),
        assignees: issue.assignees.map(assignee => assignee.login),
        createdAt: issue.created_at,
        updatedAt: issue.updated_at
      }));

      return {
        success: true,
        message: `Found ${issueList.length} issues`,
        data: {
          issues: issueList,
          repository: this.currentRepository.fullName
        }
      };
    } catch (error) {
      console.error('Failed to list GitHub issues:', error);
      return {
        success: false,
        message: `Failed to list issues: ${error.message}`
      };
    }
  }

  async getRepositoryInfo() {
    if (!this.currentRepository) {
      return {
        success: false,
        message: 'No repository selected'
      };
    }

    try {
      const status = await this.getProjectStatus(this.currentRepository.name);
      return {
        success: true,
        message: 'Repository information retrieved successfully',
        data: status
      };
    } catch (error) {
      return {
        success: false,
        message: `Failed to get repository info: ${error.message}`
      };
    }
  }

  async listCommits(args = {}) {
    if (!this.currentRepository) {
      return {
        success: false,
        message: 'No repository selected'
      };
    }

    try {
      const { limit = 10, branch } = args;
      
      const params = {
        owner: this.currentRepository.owner,
        repo: this.currentRepository.name,
        per_page: limit
      };

      if (branch) params.sha = branch;

      const { data: commits } = await this.octokit.rest.repos.listCommits(params);

      const commitList = commits.map(commit => ({
        sha: commit.sha.substring(0, 7),
        message: commit.commit.message.split('\n')[0],
        author: commit.commit.author.name,
        date: commit.commit.author.date,
        url: commit.html_url
      }));

      return {
        success: true,
        message: `Found ${commitList.length} commits`,
        data: {
          commits: commitList,
          repository: this.currentRepository.fullName
        }
      };
    } catch (error) {
      console.error('Failed to list commits:', error);
      return {
        success: false,
        message: `Failed to list commits: ${error.message}`
      };
    }
  }

  async createPullRequest(args) {
    if (!this.currentRepository) {
      return {
        success: false,
        message: 'No repository selected'
      };
    }

    const { title, body, head, base } = args;
    
    if (!title || !head) {
      return {
        success: false,
        message: 'Title and head branch are required for creating a pull request'
      };
    }

    try {
      const { data: pr } = await this.octokit.rest.pulls.create({
        owner: this.currentRepository.owner,
        repo: this.currentRepository.name,
        title,
        body: body || '',
        head,
        base: base || this.currentRepository.defaultBranch
      });

      return {
        success: true,
        message: `Pull request #${pr.number} '${title}' created successfully`,
        data: {
          number: pr.number,
          title: pr.title,
          url: pr.html_url,
          state: pr.state
        }
      };
    } catch (error) {
      console.error('Failed to create pull request:', error);
      return {
        success: false,
        message: `Failed to create pull request: ${error.message}`
      };
    }
  }

  async listPullRequests(args = {}) {
    if (!this.currentRepository) {
      return {
        success: false,
        message: 'No repository selected'
      };
    }

    try {
      const { state = 'open', limit = 10 } = args;
      
      const { data: pulls } = await this.octokit.rest.pulls.list({
        owner: this.currentRepository.owner,
        repo: this.currentRepository.name,
        state,
        per_page: limit
      });

      const prList = pulls.map(pr => ({
        number: pr.number,
        title: pr.title,
        state: pr.state,
        url: pr.html_url,
        author: pr.user.login,
        head: pr.head.ref,
        base: pr.base.ref,
        createdAt: pr.created_at,
        updatedAt: pr.updated_at
      }));

      return {
        success: true,
        message: `Found ${prList.length} pull requests`,
        data: {
          pullRequests: prList,
          repository: this.currentRepository.fullName
        }
      };
    } catch (error) {
      console.error('Failed to list pull requests:', error);
      return {
        success: false,
        message: `Failed to list pull requests: ${error.message}`
      };
    }
  }

  async getImplementationAdvice(args) {
    if (!this.genAI) {
      return {
        success: false,
        message: 'Gemini API key not configured. Set GEMINI_API_KEY environment variable.'
      };
    }

    if (!this.currentRepository) {
      return {
        success: false,
        message: 'No repository selected'
      };
    }

    const { question, context } = args;
    
    if (!question) {
      return {
        success: false,
        message: 'Question is required for implementation advice'
      };
    }

    try {
      console.log(`Getting implementation advice for: ${question}`);
      
      // Get repository context
      const repoInfo = await this.getRepositoryInfo();
      const recentIssues = await this.listIssues({ limit: 5 });
      const recentCommits = await this.listCommits({ limit: 5 });

      // Build context for Gemini
      let prompt = `You are a senior software engineer providing implementation advice for a GitHub repository.

Repository: ${this.currentRepository.fullName}
Description: ${repoInfo.data?.description || 'No description available'}
Language: ${repoInfo.data?.language || 'Unknown'}

Recent Issues:
${recentIssues.success ? recentIssues.data.issues.map(issue => `- #${issue.number}: ${issue.title}`).join('\n') : 'No recent issues'}

Recent Commits:
${recentCommits.success ? recentCommits.data.commits.map(commit => `- ${commit.sha}: ${commit.message}`).join('\n') : 'No recent commits'}

Question: ${question}

${context ? `Additional Context: ${context}` : ''}

Please provide specific, actionable implementation advice considering the repository's context, technology stack, and recent activity. Include code examples where appropriate.`;

      const model = this.genAI.getGenerativeModel({ model: 'gemini-1.5-flash' });
      const result = await model.generateContent(prompt);
      const response = await result.response;
      const advice = response.text();

      console.log(`Gemini Flash response generated successfully`);

      return {
        success: true,
        message: 'Implementation advice generated successfully',
        data: {
          question,
          advice,
          repository: this.currentRepository.fullName
        }
      };
    } catch (error) {
      console.error('Failed to get implementation advice:', error);
      return {
        success: false,
        message: `Failed to get implementation advice: ${error.message}`
      };
    }
  }

  async askClaudeToMakePlan(args, onStreamMessage = null) {
    const { prompt, workingDirectory } = args;
    
    console.log('=== CLAUDE SDK PLAN REQUEST START ===');
    console.log('Raw args received:', JSON.stringify(args, null, 2));
    
    if (!prompt) {
      console.log('ERROR: No prompt provided in args');
      return {
        success: false,
        message: 'Prompt is required for Claude planning'
      };
    }

    try {
      console.log(`✓ Prompt validated: "${prompt}"`);
      
      // Calculate target directory using DEFAULT_CODE_PATH + repository name
      let targetDirectory = workingDirectory;
      
      if (!targetDirectory && this.currentRepository) {
        const baseDirectory = process.env.DEFAULT_CODE_PATH || '/Users/reed/Code';
        const repoName = this.currentRepository.name; // Just the repo name, not full name
        targetDirectory = `${baseDirectory}/${repoName}`;
      }
      
      if (!targetDirectory) {
        targetDirectory = process.cwd();
      }
      
      console.log(`✓ Repository info:`);
      console.log(`  - Current repo: ${this.currentRepository?.fullName || 'none'}`);
      console.log(`  - Repo name: ${this.currentRepository?.name || 'none'}`);
      console.log(`  - Base directory: ${process.env.DEFAULT_CODE_PATH || '/Users/reed/Code'}`);
      console.log(`  - Target directory: ${targetDirectory}`);
      
      // Use Claude Code SDK to generate the plan with streaming
      const planResult = await this.claudeCodeManager.createPlan(prompt, targetDirectory, onStreamMessage);
      
      console.log('✓ Claude Code SDK planning completed');
      
      const response = {
        success: true,
        message: 'Claude planning completed successfully',
        data: {
          prompt,
          workingDirectory: targetDirectory,
          repository: this.currentRepository?.fullName,
          plan: planResult.plan,
          responses: planResult.responses,
          timestamp: planResult.timestamp
        }
      };
      
      console.log('✓ Response prepared with plan length:', planResult.plan.length);
      console.log('=== CLAUDE SDK PLAN REQUEST END ===');
      
      return response;
    } catch (error) {
      console.error('❌ Failed to create Claude plan:', error);
      console.error('Error stack:', error.stack);
      return {
        success: false,
        message: `Failed to create Claude plan: ${error.message}`
      };
    }
  }
}