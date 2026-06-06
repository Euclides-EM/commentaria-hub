# V6 Evaluation Report

- DB annotation evaluated: `ann_4kp7fc` (`Title Page Experiment Reviewed`).
- Current values were read from `feature_result_values.surface`; `feature_results.name` is empty for the V6 rows.
- Diagnostic matching is strict after whitespace/case/curly-quote normalization. KPI matching against `working_final_value` is a rough continuity metric, not a human-truth score.

## Diagnostic Review

- Diagnostic rows: 90
- Rows with human/preferred target: 52
- Exact target matches: 19/52
- Missing V6/current values in targeted rows: 1

| Feature | Exact | Rows | Missing |
| --- | ---: | ---: | ---: |
| Adapter Attribution | 1 | 4 | 0 |
| Adapter Description | 1 | 5 | 0 |
| Base Content | 5 | 6 | 0 |
| Base Content Description | 0 | 2 | 0 |
| Bound With | 0 | 1 | 0 |
| Date in Imprint | 1 | 1 | 0 |
| Edition Statement | 0 | 1 | 0 |
| Elements Designation | 1 | 4 | 0 |
| Enriched With | 0 | 2 | 0 |
| Euclid References | 2 | 4 | 0 |
| Explicit Language References | 1 | 3 | 0 |
| Institutions | 3 | 4 | 0 |
| Intended Audience | 0 | 2 | 1 |
| Other Educational Authorities | 2 | 2 | 0 |
| Patronage Dedication | 0 | 1 | 0 |
| Place in Imprint | 2 | 4 | 0 |
| Publisher in Imprint | 0 | 1 | 0 |
| Verbs | 0 | 5 | 0 |

| Decision Tier | Exact | Rows | Missing |
| --- | ---: | ---: | ---: |
| B_keep_manual_no_llm | 2 | 5 | 1 |
| C_llm_agreement_overrides_manual | 14 | 35 | 0 |
| D_prompt_policy_choice | 3 | 10 | 0 |
| E_single_llm_over_manual | 0 | 2 | 0 |

## KPI Continuity Check

- Working rows: 2375
- Rows mapped to DB features: 2374
- Strict matches to `working_final_value`: 1071/2374
- Unordered comma-split matches: 1090/2374
- Missing current values: 100
- Rows backed by V6 revisions: 1954/2374
- V6-revision strict matches: 924/1954

| Decision Tier | Exact | Rows | Unordered | Missing | V6 Rev Rows |
| --- | ---: | ---: | ---: | ---: | ---: |
| A_confirmed | 782 | 1292 | 782 | 10 | 974 |
| B_keep_manual_no_llm | 31 | 149 | 32 | 72 | 142 |
| C_llm_agreement_overrides_manual | 123 | 207 | 125 | 1 | 199 |
| D_prompt_policy_choice | 104 | 547 | 118 | 0 | 474 |
| E_single_llm_over_manual | 31 | 179 | 33 | 17 | 165 |

## Highest-Mismatch KPI Features

| Feature | Exact | Rows | Unordered | Missing | V6 Rev Rows |
| --- | ---: | ---: | ---: | ---: | ---: |
| Base Content | 59 | 214 | 64 | 0 | 214 |
| Date in Imprint | 82 | 198 | 82 | 0 | 0 |
| Publisher in Imprint | 77 | 192 | 78 | 6 | 192 |
| Base Content Description | 33 | 148 | 33 | 1 | 148 |
| Enriched With | 10 | 125 | 12 | 14 | 125 |
| Verbs | 77 | 171 | 80 | 27 | 171 |
| Adapter Description | 67 | 143 | 68 | 2 | 143 |
| Privileges in Imprint | 12 | 86 | 12 | 3 | 0 |
| Adapter Attribution | 113 | 169 | 114 | 0 | 169 |
| Bound With | 7 | 58 | 7 | 9 | 58 |
| Edition Statement | 19 | 65 | 20 | 2 | 0 |
| Place in Imprint | 150 | 193 | 150 | 8 | 193 |
| Elements Designation | 11 | 54 | 12 | 0 | 54 |
| Institutions | 56 | 93 | 58 | 6 | 93 |
| Other Educational Authorities | 20 | 53 | 21 | 10 | 53 |

## Current Non-V6 Revisions

These features in the KPI file are represented in the DB annotation, but their current source revision is not one of the V6 revision IDs.

| Feature | Current Revision(s) |
| --- | --- |
| Date in Imprint | v2 `71d1be08-4bb6-47d0-b31b-0c599353e775` |
| Dedication in Imprint | v2 `46fee541-0f06-4d0b-9b19-775abe81ae84` |
| Destination Language | v2 `e4d1ac46-54a8-4f86-8c7a-2c1d5cf9a101` |
| Edition Statement | v1 `718a7718-22cf-424e-a7f1-84e9cf1071c6` |
| Euclid Description | v2 `f8f833bf-6d05-46d6-8468-e2f6fe52f105` |
| Greek designation | v2 `0b9ec650-2f7b-42ef-b95b-c902d5879106` |
| Privileges in Imprint | v1 `0f63fc20-5e6b-402c-b6e7-1314707158b0` |
| Publishing Privileges | v1 `e1831df0-3c20-46d7-8df8-edc6528881f6` |

## Main Findings

- V6 is not ready as a default replacement on the reviewed rows: strict diagnostic target match is low, and risky KPI tiers `D_prompt_policy_choice` and `E_single_llm_over_manual` remain weak.
- `Base Content` improved in the diagnostic sample, but still over-expands in some rows by absorbing bound works or extra qualifiers.
- `Verbs`, `Base Content Description`, `Enriched With`, adapter fields, and `Elements Designation` need the next focused pass; most misses are span-boundary errors rather than missing extraction.
- `Date in Imprint` and `Edition Statement` did not use V6 revision IDs in the DB annotation, despite appearing in the rule file, so those cannot be judged as V6 prompt behavior from this run.
- `feature_results.name` is empty for V6 rows; consumers should read `feature_result_values.surface` for these outputs.

## Diagnostic Mismatches

| Review ID | Page/Key | Feature | Target Source | Target | V6 |
| --- | --- | --- | --- | --- | --- |
| R0008 | Alcala_1637 | Elements Designation | orig | ELEMENTOS GEOMETRICOS DE EVCLIDES | ELEMENTOS GEOMETRICOS DE EVCLIDES PHILO-SOPHO MEGARENSE SVS SEYS PRIMEROS LIBROS |
| R0051 | Amsterdam_1618 | Verbs | working | verteutscht, angefüget | verteutscht |
| R0104 | Amsterdam_1662 | Adapter Description | working | Professer Matheseos der Hooge Schoole tot Leyden | in sijn leven Professer Matheseos der Hooge Schoole tot Leyden |
| R0120 | Amsterdam_1662 | Place in Imprint | working | AMSTERDAM | t’AMSTERDAM |
| R0139 | Amsterdam_1695 | Adapter Attribution | orig | CLAAS JANSZ. VOOGHT | CLAAS JANSZ |
| R0151 | Amsterdam_1695 | Elements Designation | orig | EUCLIDIS BEGINSELEN der MEETKONST | EUCLIDIS BEGINSELEN der MEETKONST, Vervaat in 15 Boeken |
| R0170 | Amsterdam_1697 | Verbs | orig | COMPREND, REVUE, CORRIGÉE | REVUE, CORRIGÉE |
| R0171 | Amsterdam_1700 | Adapter Attribution | orig | P. CLAUDE FRANÇOIS MILLET DECHALLES | CLAUDE FRANÇOIS MILLET DECHALLES |
| R0184 | Amsterdam_1700 | Verbs | orig | EXPLIQUEZ, Reveuë, corrigée. | Reveuë, corrigée |
| R0202 | Ansbach_1610 | Adapter Attribution | orig | SIMONEM MARIUM | SIMONEM MARIUM Guntzenhu-sanum |
| R0203 | Ansbach_1610 | Adapter Description | orig | Guntzenhu-sanum Franc. Fürstlichen Brandenb: bestalten Mathematicum, vnd Medicinæ Utriusq; Studiosum. | Franc. Fürstlichen Brandenb |
| R0211 | Ansbach_1610 | Explicit Language References | orig | Griechischer, Hohe deutsche | Griechischer |
| R0222 | Antwerp_1645 | Base Content | orig | EVCLIDIS ELEMENTORVM GEOMETRICORVM LIBROS TREDECIM | EVCLIDIS ELEMENTORVM GEOMETRICORVM LIBROS TREDECIM ISIDORVM ET HYPSICLEM |
| R0308 | Arnhem_1605a | Euclid References | working | EV- CLIDIS | EV-CLIDIS |
| R0346 | Bamberg_1677 | Place in Imprint | orig | BAMBERGÆ, Franco-furtensis | BAMBERGÆ |
| R0421 | Basel_1550 | Euclid References | working | EVCLIDIS, Euclidis | EVCLIDIS |
| R0811 | Ferrara_1628 | Publisher in Imprint | orig | Franciscum Succium | Franciscum Succium Superiorum |
| R0831 | Florence_1690 | Verbs | orig | distesa, pubblicata | pubblicata, SPIEGATA, distesa |
| R0868 | Frankfurt_1674 | Adapter Description | working | REGISCURIANI E SOCIET. JESU, Gymnasio Matheseos Professoris CURSUS MATHEMATICUS | REGISCURIANI E SOCIET. JESU, Olim in Panormitano Siciliæ, nunc in Herbipolitano Franconiæ ejusdem SOCIETATIS JESU Gymnasio Matheseos Professoris |
| R0875 | Frankfurt_1674 | Institutions | working | SOC IETATIS JESU, Gymnasio Matheseos | SOCIETATIS JESU, Gymnasio Matheseos |
| R1068 | Leiden_1617 | Elements Designation | orig | xv. Boucken der Elementen Euclidis | vande xv. Boucken der Elementen Euclidis |
| R0019 | Amsterdam_1616 | Adapter Description | v5 | der stadt Leyden Landtmeter, Wijnroeyer | der stadt Leyden Landtmeter ende Wijnroeyer |
| R0021 | Amsterdam_1616 | Base Content Description | orig | Van nieus oversien, ende verbetert | Waer by ghevoecht zijn eenighe nut-ticheden, uyt de selve Boecken ghetrocken; Mitsgaders de Specien in Geometrische figueren, als ’tmaken, veranderen, ’tsamen-voughen, af-trecken, vermenichvuldighen, ende deelen |
| R0024 | Amsterdam_1616 | Edition Statement | working | Van nieus oversien, ende verbetert, Mitsgaders de bygevouchde nutticheden mette spetien in Geometrische figueren breeder verclaert ende vermeerdert deur den selven Autheur. | Van nieus oversien, ende verbetert |
| R0028 | Amsterdam_1616 | Intended Audience | working | van alle leergierighe, lief hebbers der selver vryer Conste | alle leergierighe, lief hebbers der selver vryer Conste |
| R0031 | Amsterdam_1616 | Verbs | v5 | Overgeset, verclaert, uytgeleyt, oversien, verbetert, verclaert, vermeerdert | Overgeset, verclaert, uytgeleyt, oversien, verbetert |
| R0036 | Amsterdam_1618 | Base Content Description | custom_preferred_source | Von den anfängen vnd fundamenten der Geometriæ | Die sechs ersten Bücher EVCLIDIS, Desz höchgelärten weit-berümbten, Griechischen Philosophi und Mathematici: Von den anfängen vnd fundamenten der Geometriæ |
| R0042 | Amsterdam_1618 | Enriched With | orig | mancherley auss disen Büchern gezogene nutzbarkeiten, den Speciebus inn Geometrischen figurn, als machen, verändern, zusammenfügen, abziehen, vielfältigen vnnd theilen, Per Demonstra-tiones Lineales | Dabey dann mancherley auss disen Büchern gezogene nutzbarkeiten angefüget seind, Per Demonstra-tiones Lineales |
| R0214 | Ansbach_1610 | Patronage Dedication | working | Hanss Philip Fuchss von Bimbach | Dess Edlen vnd Gestrengen Herrn, Hanss Philip Fuchss von Bimbach |
| R0224 | Antwerp_1645 | Bound With | custom_preferred_source | ISIDORVM, HYPSICLEM, Recentiores de Corporibus Regularibus, PROCLI PROPOSITIONES GEO-METRICAS | HYPSICLEM, Recentiores de Corporibus Regularibus, PROCLI PROPOSITIONES GEO-METRICAS |
| R0009 | Alcala_1637 | Enriched With | orig | comentado | comentado por LVIS CARDVCHI |
| R0098 | Amsterdam_1660 | Explicit Language References | v5 |  | Neerduyts |
| R0118 | Amsterdam_1662 | Intended Audience | orig | Aencomelingen |  |
