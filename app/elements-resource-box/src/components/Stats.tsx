import { personDisplayName } from "../utils/dataUtils.ts";
import { Row, Text } from "./common.ts";
import { useEditionsSearch } from "../hooks/useEditionsSearch.ts";
import { MARKER_4 } from "../utils/colors.ts";
import { useMemo } from "react";
import styled from "@emotion/styled";
import pluralize from "pluralize";
import { isNil } from "lodash";

const Highlight = styled.span`
  color: ${MARKER_4};
`;

type StatsProps = {
  verb?: string;
};

export const Stats = ({ verb }: StatsProps) => {
  const { items } = useEditionsSearch();

  const { editorsCount, languagesCount, citiesCount } = useMemo(() => {
    const editorsSet = new Set<string>();
    const citiesSet = new Set<string>();
    const languagesSet = new Set<string>();
    items?.forEach((item) => {
      item.editors?.forEach((editor) =>
        editorsSet.add(personDisplayName(editor)),
      );
      item.cities?.forEach((city) => citiesSet.add(city));
      item.languages?.forEach((language) => languagesSet.add(language));
    });
    return {
      editorsCount: editorsSet.size,
      citiesCount: citiesSet.size,
      languagesCount: languagesSet.size,
    };
  }, [items]);

  if (isNil(items)) {
    return;
  }

  return (
    <Row>
      {items.length === 0 ? (
        <Highlight>No matches. Try adjusting the filters or search.</Highlight>
      ) : (
        <Text size={1}>
          {verb || "Listing"} <Highlight>{items.length}</Highlight>{" "}
          {pluralize("edition", items.length)}, by{" "}
          <Highlight>{editorsCount}</Highlight>{" "}
          {pluralize("editor", editorsCount)}, in{" "}
          <Highlight>{languagesCount}</Highlight>{" "}
          {pluralize("language", languagesCount)}, from{" "}
          <Highlight>{citiesCount}</Highlight> {pluralize("city", citiesCount)}.
        </Text>
      )}
    </Row>
  );
};
