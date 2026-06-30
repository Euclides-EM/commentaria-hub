# V7 Title-Page Extraction Analysis

- Output rows: 4,360
- Completed keys: 217/217
- Feature count: 17
- Non-empty value rows: 2,291
- Empty value rows: 2,069
- Hard hallucination omissions: 11 ({'action_verbs': 10, 'editor_description': 1})
- Fuzzy-grounded rescues: 29 ({'editor_description': 4, 'bound_with': 4, 'bound_with_minimal': 4, 'base_content': 3, 'action_verbs': 3, 'dedicatee_name': 2, 'institutions': 2, 'edition_details': 2, 'educational_authorities_references': 2, 'enriched_with': 2, 'audience': 1})

## Feature Coverage
```
                        feature_id                  feature_name  rows  non_empty_values  keys_with_value  empty_keys  multi_value_keys  max_values_for_key
                      base_content                  Base Content   230               221              208           9                13                   2
              references_to_euclid             Euclid References   253               235              199          18                34                   3
                       editor_name           Adapter Attribution   226               181              172          45                 8                   3
                      action_verbs                         Verbs   479               425              163          54               115                   9
                editor_description           Adapter Description   230               162              149          68                12                   3
                   edition_details             Edition Statement   242               133              108         109                24                   3
                     enriched_with                 Enriched With   250               127               94         123                26                   5
                      institutions                  Institutions   229               100               88         129                10                   3
                bound_with_minimal          Bound With - Minimal   332               190               75         142                33                  24
                        bound_with      Bound With - Description   296               153               74         143                30                  24
              destination_language          Destination Language   225                68               60         157                 8                   2
                          audience             Intended Audience   224                63               56         161                 7                   2
                    dedicatee_name                    Dedication   217                45               45         172                 0                   1
educational_authorities_references Other Educational Authorities   268                95               44         173                20                  16
             description_of_euclid            Euclid Description   219                42               40         177                 2                   2
                   origin_language  Explicit Language References   223                46               40         177                 6                   2
                printing_privilege         Publishing Privileges   217                 5                5         212                 0                   1
```

## Policy Review Comparison
Comparison target priority: reviewer_v7_value, human_final_value, v6_target, working_final_value. `out_of_scope_no_v7_feature` marks old imprint features that this runner does not emit.
```
                     score  count
                 different     24
             span_contains     23
                 exact_set     14
out_of_scope_no_v7_feature      7
                missing_v7      4
```

### By Feature
```
                 Feature Name  different  exact_set  missing_v7  out_of_scope_no_v7_feature  span_contains
          Adapter Attribution          0          2           0                           0              2
          Adapter Description          1          2           1                           0              4
                 Base Content          0          2           0                           0              3
     Base Content Description          3          0           0                           0              1
                   Bound With          1          0           0                           0              2
            Edition Statement          1          1           0                           0              2
         Elements Designation          0          1           0                           0              5
                Enriched With          6          0           1                           0              0
            Euclid References          1          3           0                           0              0
 Explicit Language References          0          0           1                           0              1
                 Institutions          1          2           0                           0              0
            Intended Audience          2          0           1                           0              2
Other Educational Authorities          0          1           0                           0              0
         Patronage Dedication          0          0           0                           0              1
             Place in Imprint          0          0           0                           5              0
         Publisher in Imprint          0          0           0                           2              0
                        Verbs          8          0           0                           0              0
```

## Suspicious Multi-Value Extractions
```
    edition_id                  feature_name  n
  Venice_1498a          Bound With - Minimal 24
  Venice_1498a      Bound With - Description 24
  Venice_1498a Other Educational Authorities 16
   Venice_1510          Bound With - Minimal 11
     Lyon_1674      Bound With - Description 10
     Lyon_1674          Bound With - Minimal 10
   Venice_1509                         Verbs  9
    Basel_1537          Bound With - Minimal  9
   Venice_1510                         Verbs  8
Rotterdam_1661                         Verbs  7
   Venice_1505          Bound With - Minimal  7
    Basel_1546          Bound With - Minimal  6
```

## Exact Duplicate Values
```
    edition_id         feature_name         value  count
  Cologne_1591    Euclid References      EVCLIDIS      2
 Florence_1690    Euclid References     D’EVCLIDE      2
Frankfurt_1654    Euclid References      EUCLIDIS      2
     Rome_1589    Euclid References      EVCLIDIS      2
  Venice_1498a Bound With - Minimal de astrolabio      2
  Venice_1498a Bound With - Minimal      de mundo      2
```

## Hard Hallucination Omissions
- !!! llm hallucination omitted: feature=action_verbs revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a709 key=Kiel_and_Leipzig_1699 context=dataset tps and key Kiel_and_Leipzig_1699 value="einrichtung"
- !!! llm hallucination omitted: feature=editor_description revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a703 key=London_1651 context=dataset tps and key London_1651 value="Captain, Chiefe Engineer to his late Majesty"
- !!! llm hallucination omitted: feature=action_verbs revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a709 key=Rotterdam_1647 context=dataset tps and key Rotterdam_1647 value="verklaart"
- !!! llm hallucination omitted: feature=action_verbs revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a709 key=Rotterdam_1647 context=dataset tps and key Rotterdam_1647 value="verklaart"
- !!! llm hallucination omitted: feature=action_verbs revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a709 key=Rotterdam_1647 context=dataset tps and key Rotterdam_1647 value="verklaart"
- !!! llm hallucination omitted: feature=action_verbs revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a709 key=Rotterdam_1661 context=dataset tps and key Rotterdam_1661 value="verklaart"
- !!! llm hallucination omitted: feature=action_verbs revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a709 key=Rotterdam_1661 context=dataset tps and key Rotterdam_1661 value="verklaart"
- !!! llm hallucination omitted: feature=action_verbs revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a709 key=Rotterdam_1661 context=dataset tps and key Rotterdam_1661 value="verklaart"
- !!! llm hallucination omitted: feature=action_verbs revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a709 key=Rotterdam_1681 context=dataset tps and key Rotterdam_1681 value="verklaart"
- !!! llm hallucination omitted: feature=action_verbs revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a709 key=Rotterdam_1681 context=dataset tps and key Rotterdam_1681 value="uit-geleydt"
- !!! llm hallucination omitted: feature=action_verbs revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a709 key=Utrecht_1647 context=dataset tps and key Utrecht_1647 value="verklaart"

## Fuzzy Grounding Examples
- warning: llm fuzzy-grounded near hallucination: feature=dedicatee_name revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a717 key=Florence_1690 context=dataset tps and key Florence_1690 value="AGL’-ILLUSTRISSIMI SIG. DELL’ACCADEMIA DE’ NOBILI" source_value="AGL’-ILLVSTRISSIMI SIG. DELL’ACCADEMIA DE’ NOBILI"
- warning: llm fuzzy-grounded near hallucination: feature=editor_description revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a703 key=Lyon_1660 context=dataset tps and key Lyon_1660 value="Camberiensis Societ. IESV" source_value="Camberiensi Societ. IESV"
- warning: llm fuzzy-grounded near hallucination: feature=institutions revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a714 key=Lyon_1660 context=dataset tps and key Lyon_1660 value="Camberiensis Societ. IESV" source_value="Camberiensi Societ. IESV"
- warning: llm fuzzy-grounded near hallucination: feature=bound_with revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a718 key=Lyon_1690 context=dataset tps and key Lyon_1690 value="Theodosii sphærica, Sectiones Conicas, Arith-maticam, Trigonometriam, Algebram, & refutationem Hyptheseon Cartesianarum" source_value="Theodosii sphærica, Sectiones Conicas, Arith-meticam, Trigonometriam, Algebram, & refutationem Hyptheseon Cartesianarum"
- warning: llm fuzzy-grounded near hallucination: feature=bound_with_minimal revision=b58a2f9b-dc7d-459f-8953-4c3cf08aa114 key=Lyon_1690 context=dataset tps and key Lyon_1690 value="Arith-maticam" source_value="Arith-meticam"
- warning: llm fuzzy-grounded near hallucination: feature=editor_description revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a703 key=Paris_1565 context=dataset tps and key Paris_1565 value="lecteur ordinaire\ndu Roy ès Mathema-\ntiques, en l’vni-\nersité de\nParis" source_value="lecteur ordinaire\ndu Roy és Mathema-\ntiques, en l’vni-\nuersité de\nParis"
- warning: llm fuzzy-grounded near hallucination: feature=institutions revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a714 key=Paris_1565 context=dataset tps and key Paris_1565 value="l’vni- ersité de\nParis" source_value="l’vni-\nuersité de\nParis"
- warning: llm fuzzy-grounded near hallucination: feature=bound_with revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a718 key=Paris_1573 context=dataset tps and key Paris_1573 value="Πυθαγόρας σοφὸς εὗρε, Πλάτων δ’ ἀρίδηλ’ ἐδιδάξεν" source_value="Πυθαγόρας σοφὸς εὗρε, Πλάτων δ’ ἀρίδηλ’ ἐδί-\nδαξεν"
- warning: llm fuzzy-grounded near hallucination: feature=bound_with revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a718 key=Paris_1609 context=dataset tps and key Paris_1609 value="les quatorze & quinziesme d’Ipsicles Alexandrin" source_value="les quatorze & quinziesme d’Ipsicles\nAlexandrie"
- warning: llm fuzzy-grounded near hallucination: feature=bound_with_minimal revision=b58a2f9b-dc7d-459f-8953-4c3cf08aa114 key=Paris_1609 context=dataset tps and key Paris_1609 value="les quatorze & quinziesme d’Ipsicles Alexandrin" source_value="les quatorze & quinziesme d’Ipsicles\nAlexandrie"
- warning: llm fuzzy-grounded near hallucination: feature=edition_details revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a720 key=Paris_1609 context=dataset tps and key Paris_1609 value="Traduits & restitués à leur ancienne breuété, selon l’ordre de Theon" source_value="Traduits & restitués à leur ancienne breueté, selon\nl’ordre de Theon"
- warning: llm fuzzy-grounded near hallucination: feature=educational_authorities_references revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a716 key=Paris_1609 context=dataset tps and key Paris_1609 value="Ipsicles Alexandrin" source_value="Ipsicles\nAlexandrie"
- warning: llm fuzzy-grounded near hallucination: feature=bound_with_minimal revision=b58a2f9b-dc7d-459f-8953-4c3cf08aa114 key=Paris_1610 context=dataset tps and key Paris_1610 value="les quatorze & quinziesme d’Ipsicles Alexandrin" source_value="les quatorze & quinziesme d’Ipsicles\nAlexandrie"
- warning: llm fuzzy-grounded near hallucination: feature=educational_authorities_references revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a716 key=Paris_1610 context=dataset tps and key Paris_1610 value="Ipsicles Alexandrin" source_value="Ipsicles\nAlexandrie"
- warning: llm fuzzy-grounded near hallucination: feature=editor_description revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a703 key=Paris_1654a context=dataset tps and key Paris_1654a value="R. P., de la Compagnie de IESVS" source_value="R, de la\nCompagnie de IESVS"
- warning: llm fuzzy-grounded near hallucination: feature=bound_with revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a718 key=Rome_1655 context=dataset tps and key Rome_1655 value="Ex maioribus CLAVII Comentariijs" source_value="Ex maioribus CLAVII Com-mentarijs"
- warning: llm fuzzy-grounded near hallucination: feature=bound_with_minimal revision=b58a2f9b-dc7d-459f-8953-4c3cf08aa114 key=Rome_1655 context=dataset tps and key Rome_1655 value="Comentariijs" source_value="Com-mentarijs"
- warning: llm fuzzy-grounded near hallucination: feature=base_content revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a705 key=Rotterdam_1632 context=dataset tps and key Rotterdam_1632 value="De ses eerste Boecken EVCLIDIS, Van de beginselen ende fundamenten der Geometrie" source_value="De ses eerste Boecken EVCLIDIS, Van de beginselen ende fondamenten der Geometrie"
- warning: llm fuzzy-grounded near hallucination: feature=action_verbs revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a709 key=Rotterdam_1647 context=dataset tps and key Rotterdam_1647 value="breeder verklaart" source_value="breeder verklaert"
- warning: llm fuzzy-grounded near hallucination: feature=enriched_with revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a708 key=Rotterdam_1661 context=dataset tps and key Rotterdam_1661 value="Mitsgaders de by-gevoeghde nuttigheden, met de specien in Geometrische figuren breeder verklaart ende vermeerdert" source_value="Mitsgaders de by-gevoeghde nuttigheden, met de specien in Geometrische figuren breeder verklaert ende vermeerdert"
- warning: llm fuzzy-grounded near hallucination: feature=editor_description revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a703 key=Rouen_and_Paris_1677 context=dataset tps and key Rouen_and_Paris_1677 value="D. Mathematicien" source_value="n Mathematicien"
- warning: llm fuzzy-grounded near hallucination: feature=base_content revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a705 key=Strasbourg_1570a context=dataset tps and key Strasbourg_1570a value="EVCLIDIS Propositiones Elementorum" source_value="EVCLIDIS Propositiones. Elementorum"
- warning: llm fuzzy-grounded near hallucination: feature=base_content revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a705 key=Strasbourg_1599 context=dataset tps and key Strasbourg_1599 value="ΕΥΚΛΕΙΔΟΥ ΣΤΟΙΧΕΩΝ ΠΡΩΤΟΝ. EVCLIDIS ELE-\nMENTVM primum" source_value="ΕΥΚΛΕΙΔΟΥ ΣΤΟΙΧΕΙΩΝ ΠΡΩΤΟΝ. EVCLIDIS ELE-\nMENTVM primum"
- warning: llm fuzzy-grounded near hallucination: feature=enriched_with revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a708 key=Urbino_1575 context=dataset tps and key Urbino_1575 value="da li riueduti" source_value="da lui riueduti"
- warning: llm fuzzy-grounded near hallucination: feature=action_verbs revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a709 key=Utrecht_1647 context=dataset tps and key Utrecht_1647 value="breeder verklaart" source_value="breeder verklaert"
- warning: llm fuzzy-grounded near hallucination: feature=edition_details revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a720 key=Venice_1505 context=dataset tps and key Venice_1505 value="addita sub nec non plurima subuersa et prepostere: voluta in Campi interpretatione: ordinata digesta et castigata sunt" source_value="addita sub nec non plurima subuersa et prepostere: voluta in Campani interpretatione: ordinata digesta et castigata sunt"
- warning: llm fuzzy-grounded near hallucination: feature=audience revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a715 key=Venice_1510 context=dataset tps and key Venice_1510 value="quicumque ad mathematic-\nicam substantiam aspirant" source_value="quicumque ad mathemat-\nicam substantiam aspirant"
- warning: llm fuzzy-grounded near hallucination: feature=dedicatee_name revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a717 key=Venice_1543 context=dataset tps and key Venice_1543 value="Professore di tal Scientie Nicola Tartalea, BRISCIANO" source_value="Professore di tal Scientie Nicolo Tartalea, BRISCIANO"
- warning: llm fuzzy-grounded near hallucination: feature=action_verbs revision=8c74a0d1-b4d5-4b90-8b44-0a6771c2a709 key=Vienna_1694 context=dataset tps and key Vienna_1694 value="eingesrichtet" source_value="eingerichtet"