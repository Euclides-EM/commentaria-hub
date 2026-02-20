import { VersionControlService } from '@hub-api';

export interface RepoPullResponse {
  success: boolean;
  branch: string;
}

export interface RepoPrResponse {
  success: boolean;
  branchName: string;
  prUrl: string;
  prNumber: number;
}

export const pullRepo = async (): Promise<RepoPullResponse> => {
  console.log("Pulling repository");
  const result = await VersionControlService.postVersionControlPull();
  return {
    success: result.success ?? false,
    branch: result.branch ?? "",
  };
};

export const createPullRequest = async (): Promise<RepoPrResponse> => {
  console.log("Creating pull request");
  const result = await VersionControlService.postVersionControlPush();
  return {
    success: result.success ?? false,
    branchName: result.branch ?? "",
    prUrl: result.pr?.url ?? "",
    prNumber: result.pr?.number ?? 0,
  };
};
