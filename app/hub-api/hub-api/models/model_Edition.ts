/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { model_EditionSubjectCategory } from './model_EditionSubjectCategory';
import type { model_EditionTitlePageStatus } from './model_EditionTitlePageStatus';
import type { model_EditionVisualElement } from './model_EditionVisualElement';
export type model_Edition = {
    additionalContent?: Array<string>;
    bibliography?: Array<string>;
    books?: Array<number>;
    /**
     * Print-only
     */
    cities?: Array<string>;
    colophon?: string;
    colophon_EN?: string;
    corpus?: Array<string>;
    diagramCropsAvailable?: boolean;
    editor?: Array<string>;
    format?: number;
    frontispiece?: string;
    frontispiece_EN?: string;
    hasDiagrams?: boolean;
    imprint?: string;
    imprint_EN?: string;
    /**
     * Elements (both)
     */
    isElements?: boolean;
    /**
     * Manuscript-only
     */
    isManuscript?: boolean;
    key?: string;
    languages?: Array<string>;
    manuscriptClass?: string;
    manuscriptSubclass?: string;
    manuscriptYearFrom?: number;
    manuscriptYearTo?: number;
    notes?: string;
    publisher?: Array<string>;
    reprintOf?: string;
    shortTitle?: string;
    shortTitleSource?: string;
    readonly subjectCategories?: Array<model_EditionSubjectCategory>;
    title?: string;
    titlePageStatus?: model_EditionTitlePageStatus;
    title_EN?: string;
    ustcId?: string;
    verified?: boolean;
    visualElements?: Array<model_EditionVisualElement>;
    readonly visualElementsTypes?: Array<string>;
    volumes?: number;
    wardhaughClassification?: string;
    year?: string;
};
