import * as githubTools from './github-tools.js';
import * as gitTools from './git-tools.js';
import * as projectTools from './project-tools.js';

export const getAllTools = () => {
  return [
    githubTools.createGitHubIssue,
    githubTools.updateGitHubIssue,
    githubTools.listGitHubIssues,
    githubTools.closeGitHubIssue,
    
    gitTools.gitStatus,
    gitTools.gitCommit,
    gitTools.gitPush,
    gitTools.gitPull,
    gitTools.gitBranch,
    gitTools.gitLog,
    
    projectTools.listProjects,
    projectTools.selectProject,
    projectTools.projectStatus,
    projectTools.createProject
  ];
};

export const getToolByName = (name) => {
  const allTools = getAllTools();
  return allTools.find(tool => tool.name === name);
};

export const getToolSchemas = () => {
  return getAllTools().map(tool => ({
    name: tool.name,
    description: tool.description,
    parameters: tool.parameters
  }));
};

export {
  githubTools,
  gitTools,
  projectTools
};