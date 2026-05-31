/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { model_Model } from "./model_Model";
import type { model_ModelTrainingStatus } from "./model_ModelTrainingStatus";
export type model_ModelTraining = {
  backend?: string;
  readonly created_at?: string;
  epochs?: number;
  gpu_farm_host?: string;
  readonly id?: string;
  model?: model_Model;
  remote_run_dir?: string;
  status?: model_ModelTrainingStatus;
  status_details?: Record<string, string>;
  readonly updated_at?: string;
};
