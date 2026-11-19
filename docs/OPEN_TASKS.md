# Add Jesuits information

* Go over the data in https://jsn.soohwalee.com/ and the additional raw data Volker sent me, in here: https://uni-wuppertal.sciebo.de/s/oLCymSpp6gCMCkx
* Check how the data can be integrated into the existing infrastructure.
* Implement or create new tasks for integration.

# Editions prompt based classification

We already integrated an LLM based subject classification of the editions, especially those that are not Euclid's Elements. The classification is based on the title pages transcriptions. We run it on the editions a few month ago. Now, we need to:
* Review the categories and prompt. 
* Adapt the prompt if needed and conduct small experiments with limited number of editions, to test the adaptations.
* Once we are satisfied, re-run the classifications on all of the editions to update the old classifications and classify the new editions.

# Add manuscripts to the catalog [IN PROGRESS: Liri]

This is a request received from multiple scholars at CIRM, Oct 2026.

Currently, the catalog only supports printed editions, but we want to add support for manuscripts as well. This requires diffrent approach to the material.

The work to add the manuscripts was already done in https://github.com/Euclides-EM/commentaria-hub/pull/73, but since then we made multiple changes. Therfore, we need to:
* Align the [PR](https://github.com/Euclides-EM/commentaria-hub/pull/73) to the new code.
* Fix according to the review in the [document](https://docs.google.com/document/d/1zc2cSq8M2GXC5enRmirEuZQwwhzJVEqzo6yUsxO3xIk/edit?usp=sharing)

Then, we need to add a few manuscripts:
* Euclid in the Portuguese Empire corpus [link](https://collectionguides.universiteitleiden.nl/resources/ubl248#) 
* Leiden mathematical manuscript collection [link](https://collectionguides.universiteitleiden.nl/resources/ubl248)
* Mia's collection (https://drive.google.com/drive/folders/1NkdGJ5lOqpG632vPmRkGXLHMKJ9mPq_J?usp=drive_link)
* Max's collection
* Possibly: collections or shelfmarks from Catherine Goldstein & Pierre Ageron (I heard his [lecture](https://hidim2026paris.sciencesconf.org/resource/page/id/3))

# Detect reprints

We integrated in the UI a button the allows us to detect and review reprints. Use this functionality and review all the existing suspected reprints. In addition, at some point I created a small list of suspected reprints, review those as well, and determine which are reprints. For each edition identified as a reprint, mark it accordingly in the UI, or directly in the CSV.

The list I prepared:
* 1496 and bib-217
* Bib-224, bib-225
* OM1R2Y a reprint of something
* Bib-46, bib-47, bib-48, bib-49, bib-50
* Bib-21, bib-129
* 0NKS71, YBCWZV

# Integrate name variation for editors

Currently, when the user searches for specific editor, author or publisher, there is only one spelling that can be used to find them. We need to allow for variation on spelling since: (1) we need to support search by initials and (2) for many of these individuals, especially the early ones, there were a few options to spell their name. In the past, we created a list based on a few identity repositories, such as VIAF. It is located in the document "Editors and Publishers in DB - April 2026". However, we also need to write crawlers that automatically retrieve the variations. The acceptance criteria of this task is to:
* Add a name variation support to the frontend and backend that uses a CSV stored in [ocrflow/store/items_metadata](../ocrflow/store/items_metadata).
* Add a pipeline that we can run and fills the variation mapping based on crawlers on identity repositories.
* Add the relevant information as a skill and to AGENTS.md.
* Some effort already made at https://github.com/Euclides-EM/commentaria-hub/pull/101

# Detect and fix invalid diagram crops

We found out that some of the crops are not available in the server's UI ("failed to load" error). For example: 
* [London_1680](https://elements-resource-box.netlify.app/diagrams?key=London_1680%E2%80%9381)
* [Würzburg_1661](https://elements-resource-box.netlify.app/diagrams?key=W%C3%BCrzburg_1661)
We need to:
* Add an automatic process that detects invalid crops and runs weekly as a Github action. If invalid crops are detected, an email should alert Mia, Denisa and Lennart.
* Fix all existing invalid crops.

# Obtain missing facsimiles

* Coordinate with the project's PIs and finalize the list of editions we want to transcribe.
* For each edition that we want to transcribe and we do not have a facsimile with proper copy rights (either public domain or NC): (1) try to find a facsimile online and if you do not succeed (2) obtain it from a library.
* If we need to pay for the facsimile, communicate the costs to the PIs.

Note: Max already started to handle it in the past, he created a report (in Mia's email). In addition, Mia made a tiny list as well at some point, I'm not sure how up to date it is: Uppsala_1744, Halle_1758 (for now we use the 1818 reprint [56DVN2]), Naples_1810 (for now we use the 1818 reprint 1814 [HZ5UVJ]), Lund_1855

# Transcribe additional editions

Continue curating the editions' transcription, according to [TRANSCRIPTION_STATUS.md](TRANSCRIPTION_STATUS.md).

# Cost estimate for LLM Corrections

Create a cost estimate report for the PIs to review.

# Integrate third party transcriptions and indices

There are facsimiles that were already indexed or have been fully transcribed. 
* Create a list of such transcriptions/indexes in a new markdown file in this folder. Note for each transcription/index any metadata information that can be recorded. 
* If you find transcription/index that are accurate and add value, find a way to integrate them. 

Note: in the past, I found these facsimiles:
* [London 1570](https://diglib.hab.de/drucke/27-9-geom/start.htm?image=00005)
* [Multiple editions](https://app.transkribus.org/en/sites/noscemus/browse) - these seem to be of low quality, but we can audit them.

# Search by part

One of the requests we got is the ability to search in a "part" of the edition. For examole "Scholia that includes the word X" or "passage in Book II that discuss Y." This feature should utilize both the indexing feature and any search functionality we integrate, either the textual search available todayy, or the normalized/semantic one we will integrate in the future. Moreover, any extention to the search functionality should support the search-by-part option.

# Search based on normalized text

Currently, users can search in the transcriptions' text using exact match. So, for example, if the transcription includes the word "diuision" but the user searched for "division", the search will not work. We need to devise a plan to add normalized search, that allows the user to find relevant passages even if there are no exact match. Initially, the search functionality can use simple options, like rank search based on some approximation algorithm, that takes into account common historical spelling variation, such as long-s, u vs. v, etc. We also need to integrate a "smart" data model and accompanying infrastructure that will allow us to use algorithms that allow fast ranked search.

This is a large feature, that can be addressed in stages. While we can start with textual search, later we can add semantic search, based on Huma-Num's Ollama server or other AI providers. We might want to use the existing feature extraction functionality already integrated.

# Integrate NER for individual, places and concepts

Enhance the transcriptions, on the TEI level, with a named entity recognition (NER) layer that allows to search, highlight and extract passages that contain names of individuals (such as "Theon"), places ("Egypt") and mathematical terms or concepts ("infinity", "Pythagoras theorem").

This might be related to the search functionality and to the existing feature extraction functionality, already integrated in the system.

We can implement this feature in stages, starting with a few categories, and enhancing the feature slowly.

# Correlate the propositions to Heiberg's Elements skeleton

For each proposition identified in the transcribed editions, add an alternate index that correlate it to Heiberg's proposition numbering. Use an LLM to accomplish this task.

# Per-proposition view for transcribed editions

Add a new view in the UI that allows the user to compare a single proposition across edition. Allow the users to select the filtering criteria, so they can either specify a closed list of editions or a condition such as "German editions printed before 1650". The users sholud also be able to select wether they want to identify the propositions according to the structure in the edition, or according to Heiberg's Elements skeleton. 

While this feature includes a frontend component, it might also require a creation of alternate indexing pr an extension of it.

# Integrate diagrams work

Denisa is working on the diagrams. The output of her work should be integrated in the platform.

# Huma-Num Backup

Currently we have a backup mechanism that allows creation of a backup on demand through the UI/API and stores the backup both in Mia's personal Google Drive and on the server's volume. We had a job that run weekly to create the backup. This job zipped all te data except of the raw facsimiles and diagram crops, which are exceptionally heavy, and uploaded the zip. However, we changed some of the settings after we hit space limits. The diagrams and facsiilies are moved using SCP to Mia's hard drive, with the process described in [FACSIMILE_MEDIA_BACKUP.md](FACSIMILE_MEDIA_BACKUP.md)

We need to:
* Change the logic of the scheduled backup, so instead of a gigantic ZIP we will use a library such as Restic (see [this](codex://threads/01a1b012-7eda-7fb0-baf5-621217a3b2bb) and [this](codex://threads/01a147e1-c9f7-7ef1-af0d-a137c5932442) Codex threads).
* Restore the weekly backup, and alert Mia's email if the backup fails.
* Possibly enhance the monitor functionalities and the logic for facsimilies/diagrams backup.

# Tech Debt: Separate imprint and title page features input

Currently, the imprint and title page features are separated using a hack, fix it and work according to best practices: https://github.com/Euclides-EM/commentaria-hub/pull/85

# Integrate social/email login

This is a request received from multiple scholars at CIRM, Oct 2026.

Currently, all users are anonymous, except a small number of users that were hard coded in a list in [httpwrapper/auth.go](../ocrflow/pkg/httpwrapper/auth.go) and are authenticated using a Github token. This creates a friction when adding new users. To facilitate wider community engagement, we need to integrate registration and authentication flows using email, with 2FA. We should use a third party, such as Firebase. For the registration flow, each registration request should be approved by one of the project's administrators (currently: Mia, Denisa, Lennart) 

# Integrate indexes in the catalog filters

This is a request received from multiple scholars at CIRM, Oct 2026.

Currently, we support indexes for transcribed editions. Curated editions have on top of that also alternate indexes. We need to add the indexes, to some extent, to the catalog filters, allowing, for example, for users to search for "editions that contain scholia in Book I", "editions that have practical usage sections for II.5" etc.

# Enhance the metadata available to anonymous users

This is a request received from multiple scholars at CIRM, Oct 2026.

Currently, anonymous users can view only some of the metadata available on the edition. Many details are only available in the edit mode, which is accessible only to authenticated users. We need to enhance it, so all users will be to see the metadata. In particular, scholars at the CIRM conference (Oct 2026) asked to see shelfmarks.

# Download as bibtex

This is a request received from Catherine Goldstein, Oct 2026.

Allow users to download bibtex files from the catalog, based on their search.

# Add Russian editions

This is a request received from Andrei Rodin, Oct 2026.

Andrei discussed with Mia some Russian editions. For some he obtained the PDFs. While the PDFs might not be with appropriate copyrights, we should at least integrate the metadata and note that Andrei holds some facsimiles.

# Integrate Catherine's corrections to the catalog

This is a request received from Catherine Goldstein, Oct 2026.

Catherine has sent some corrections to the catalog, address them.

# Filters and/or logic is unclear

This is a request received from Catherine Goldstein, Oct 2026.

Currently, we only support "or" logic in the catalog. So, if a user searches for editions that contain certain books of the Elements, and specifies "Book I", "Book VII" they will get all editions that include either books. However, ths is not intuitive and does not allow users to search for editions that include, for example, both Book I and Book VII. We need to integrate an intuitive, user-friendly, was to conduct such search, without complicating the UI.

# Add acknowledgment page to the platform

This is a request received from Vincenzo, Oct 2026.

Add an acknowledgment page that mentions the all the researchers and funding sources of the project, including scholars that contributed by sending catalogs, feedbacks, etc.

# Add Spanish and British editions [Reviewed, ready for action]

This is a request received from Vincenzo, Oct 2026.

A british scholar sent Vincenzo a list of Spanish and British editions to be added to the catalog (see email titled "Spanish anf British"). We need to both add them, and add the scholar's name to the acknowledgment page.

Summary from the sources is at [this doc](https://docs.google.com/document/d/1o31L6fSTd4iv-RzofF5B-I3Sw-9Aum_7/edit?usp=sharing&ouid=107005912857829665978&rtpof=true&sd=true).

# Add Scandinavian editions

This is a request received from Gorm Tortzen, Oct 2026.

Gorm Tortzen has sent me a list of Scandinavian translation to be added to the catalog. We need to both add them, and add his name to the acknowledgment page. The editions are listed here and can be found by searching for Euclides (with the Scandinavian languages spellings): [link](https://skandinaviske-oversaettelser.net/)
