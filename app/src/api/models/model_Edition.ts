/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { model_EditionShelfmark } from './model_EditionShelfmark';
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
    shelfmarks?: Array<model_EditionShelfmark>;
    shortTitle?: string;
    shortTitleSource?: string;
    title?: string;
    title_EN?: string;
    ustcId?: string;
    verified?: boolean;
    visualElements?: Array<model_EditionVisualElement>;
    volumes?: number;
    year?: string;
};

