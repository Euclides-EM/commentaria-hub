/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { search_OrderByOption } from './search_OrderByOption';
import type { search_Range } from './search_Range';
export type search_Query = {
    fields_filter?: Record<string, Array<string>>;
    filter_includes?: Record<string, boolean>;
    limit?: number;
    offset?: number;
    order_by?: Array<search_OrderByOption>;
    range_filter?: Record<string, search_Range>;
    text_search?: string;
    text_search_fields?: Array<string>;
};

