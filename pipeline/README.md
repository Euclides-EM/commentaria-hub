
# What is this doc?

It's a live, technical design doc that includes both the design and the workplan.
It's technical in part, and very practical in other parts. Like our beloved Elements, it can be read in a non-linear way. 
TODO - share a drive link, not a pdf... Add manicules!!

# Overview
This project's goal is...

The pipeline "steps" are all done together - so, we create a skeleton pipeline and enrich it. So, the pipeline design does not signal a work plan, but the pipeline... The work plan is layed out later.

# Design Principles
There are a few principles that drives all of the choises in this pipeline design. This is out מסגרת רעיונית and they are derived from the requirments of the project. Let's spell them out.

Not invent the wheel, but extend it. If someone did it before us, let's use it! If there is a standard that is esablished, let's use it. If it's not a perfect fit, we try to enhance and extend, not re-create from scratch. Why? first - to save time; Second - so we can contribute to the field: the final datasets, our code and even more abstract methodology.

Flexibility and extendability. We want it to be super easy to add new editions to the corpus, but not only new editions, but also new inputs - e.g., if we get a second digital facsimilie, a new groud truth dataset, a new model that we want to use. This means a few things. First, adding should be automated as much as possible and whatever that can't be automated should be clearly documented how to do it. Second, nothing should be too specific. Ofc the models and all should be per-edition, but the overall processes that generate the model should be non specific and automated. Third, reproducible (as much as possible) the automated processes should be grounded in something that can be re-run, no obscure dataset that we get, tune the model and then we do not know how we got to the model.

The sacred triangle: rule-based, LLM and research-in-the-loop. We use LLMs extensively for any task that we would generally give a human. However, to prevent the typical disadvantages of LLMs (helucinations, etc.) we remember to (1) KISS - keeping it simple, if a task can be rule-based, if it makes more sense to use some non LLM algorithm, there is no reason why not. E.g., there are advanced error-correction algorithms that have nothing to do with LLMs. Second - research in the loop - the pipeline is transperant and interactive. Thus, the researcher can either override the LLM decision or re-direct it, changiung it's prompts or the places it is being employeed. The researcher can always review the automatic processes making and in critical sections the researcher is required to do so. The interactivity allows a iterative processe, where the machine does its thing, the user reviews and takes action and the automatic processes are changed and re-employed.

# The Pipeline High Level Overview

Why? Using the Gallic(orpor)a workflow.
The pipeline's process is to establish an organied, documented, clear & reproduciable way to re-create the entire process.

Main Parts:
- Obtaining the raw data and metadata
- Segmentation and transcription
- TEI Encoding
- Post-processing and Quality Assurance - integration of LLMs + spell-checking
- Publishing
- Search functionality - includes indexing + dictionary building
- Diagrams
- Additional editions

# Getting the editions raw data and metadata

## What is this step?
This step is about gathering the raw data (PDFs of digitized documents) and their metadata (information about these documents, such as title, author, date, etc.) into a structured dataset (CSV). These are the editions we actually want to process.

In terms of how we do it, we use an LLM with research-in-the-loop approach... (copy the mail I sent to Angela and Max). The dataset can be found here [add link to CSV without the additional columns].

I could not locate all the editions and suggested alternatives, see appendix A. (the doc I sent to Vincenzo)

Currently, all raw data is stored here [add a link to facsimile repo], and uses Github LFS [add link] - but it needs to move to HumaNum.

## Current Status
Essentially, this step is more or less done for an MVP. But, we can (and should) make it more qualitative. 
Thus, the following tasks:
- Going over Max and Angela's work and integrating it here. I have a meeting with Max to discuss it.
- Moving the raw data to HumaNum volume.
- Possibly, additional facsimile will be added later - either more editions we want to add to the project OR, if we can have multiple facsimiles for editions it can make the OCR results better, since we can correlate. 
If Angela and Max are open to it, I can use Max's help with it.

# Additional Inputs for the OCR

## What is this step?

To make our OCR better, we use both datasets and models that are created by not-us. This step includes compiling these datasets and base models. The idea is that we use the base models and only tune them based on the datasets (those we select as relevant and those created by us that we reviewed and deemed in good quality).

The datasets we are using are... TODO

In addition, some of the PDFs are searchable, I created a script TODO that specifiees for each PDF if it's searchbale or not. If it is, it means that something (probably a machine...) trascribed it, these results can be correlated to the OCR results we created and compared. With these datasets, if we see big discrepancies it means that 

The models that currently I'm using as base are... From some PoC I believe __ gives better results, but it's not in stone...

## Status

As with the step before, we are more or less done for MVP.

# Segmentation and transcription
## What is this step?
This step is really the heart of the entire thing, and the most complex one. This is what creates from the facsimile the transcribed content. The facsimiles are processed to extract their layout and text.

The standard way to do it to use a model to segmenting and then another model to transcribe. The idea is that for each page in the facsimile two seperate models run. The first is identifying the diff zones in the page - e.g., the diagrams, catchwords, paragraphs, headers, etc. Then, another model runs and extracts the lines. The another model runs - these one is the transcription model, that for each line, figures out what actually is written in it. Usually, the models that run are some shelf model that is fine tuned. What des it mean? We take the model and fine tune it with additional datasets that are ground truth. How are these obtained? Either from some other project, or from manually correcting a few pages that the model processed and re-training the base model. Usually, after all that, a set of scripts runs to do some post fixes. These can be simple scripts - e.g., if the running header is always "Euclid Liber <roman numeral>" that is going up in jumps of 1 (e.g. "Euclid Liber I" is p. 30-50, "Euclid Liber I" p. 51-43, ...) and suddenly between two pages of "Euclid Liber IX" there is a "Euclid Liber IXI", it's probably a mistake, and can be easily corrected. The post scripts can also be more complex, e.g.., by employing a dictionary check with external dictionaroes that are also sensitive to common mispless in the transcribed facsimilie. Lately, some are also adding LLM based post fixes, that asks LLM to correct the results.

Our appriach builds on this standad approach, but it is different. Let's say we have edition(s) we want to OCR. In this case, the researchers launches a command line interface and gives as an input the list of editions she wants to OCR - the link to the facsimilie + the metadata (year, lang, etc.). Then, the software goes over the base datasets and models and uses a set of rules to decide for each edition (1) what is the best base segmentation model (2) what is the best base transcription model and (3) what are the datasets relevant for the tuning. After these pieces of info are chosen, the researcher is prompted the results for each edition & reason for each choice. The user can either approve the choice or change it by adding exception rules that will be applied in the next time the flow will run. For example, let's say this step runs on the edition ..., then the base transcription model ... will be chosen, since the script is ... and the dataset ... will be chosen and not ..., since it's not relevant for the time. Note that the base model might be a model we trained before. E.g., if another reprint of Clavius is added, and we already completed the prev Clavius models, then it'll be used.

Then, the selected models is tuned accirding to the choses of datasets and new models are created and stored, including all of their metadata. Note that the model is per-edition. Then, the model runs on the edition and the post scripts are running. These scripts might call each other and work in an iterative way (sort of a discussion between machines) - so, for example... If it is relevant, the post-scripts might prompt a question for the researcher. For example, if the confidence is very low in something it might be sent to an LLM and if the LLM is not sure, it might be ascelated to the user. The process can be stopped and started. It might be a long process, so the ability to stop it - even to review in the moddle - is supported.

After all of the post-scripts run (or if it's stopped), the user gets both the full audit of which scripts where invoked and when - sort of a "transcription" of the discussion between the machines - and the actual OCR. That OCR is stored with a version num and with the metadata of when, by who, with ehich models, full audit of the ping-pong of the post-scripts. The researcher has a few choises:
* She can accept it as the "best" option, mark it as the edition's "official" OCR.
* She can 
* She can request to run "verification" on all of the pages and view the pages that are deemed as most or least good-OCR. We will implement a few such verification logics, some use LLM.
* She can review the resulting pages. If she wants, she can create a new dataset or add to an existing one pages she deems ground-truth.
The user can view  run on the editions and we get the basic segmented and transcribed.

The idea is that to create the OCR of an edition facsimile we use a few builing blocks that are combined in an iterative way. The building blocks are:
* Transcribtion and segmentation models

The first thing to remember here is that we are working on very diff editions. There are grand diffrences between a text printed in Basel in 1500 and one printed in Sweden in 1800, a one-size-fits-all segmentation and transcription will be too crude. In addition, we want flexibility. Meaning, we want it to be easy to add one more edition to the corpus and we want that additional ground truth datasets - either made by manual/LLM corrections OR by obtaining additional relevant datasets from other projects - will be possible and automatic. The last thing we want is to not invent the wheel - so, we use base models that were developed by others and tune them.

So what is our approach? It is built on two basic principles: extensive and varied use of LLM + "research in the loop". We use LLM in three ways: first, as a choser - in some cases in the process we need to make decisions and instead of a human making them, 

However, 

For the segmentation, we use the SegmOnto vocabulary, as described in this[add link] article. This vocabulary is specialized for early modern books, it is standard, so others are using it and it allows us to use ground truth datasets from the community as well as contribute to it once we are done.

So.. how do we actiolly obtain the  

For both models, we use the following approach:
- First, we use a base model that we tuned specifically for the edition (more on it later) and produce layout+transcription
- Then, we run a set of scripts that correct the transcription LLM instruction (again, specific per edition) that prompt an engine to correct the layout+transcription
- At the end, we 

So, the approach we take is that we are not developing a model, but a flow that outputs a reproducible procedure that generates models. 
The flow's steps are:
1. Go over the corpus list, which include one entry per edition we want to transcribe. Go over the base datasets and models and use a set of rules to decide for each edition (1) what is the best base segmentation model (2) what is the best base transcription model and (3) what are the datasets relevant for the tuning. After these pieces of info are chosen, the researcher can view the automatic decision & reason for each choice, and add exception rules that will be applied in the next time the flow will run. For example, let's say this step runs on the edition ..., then the base transcription model ... will be chosen, since the script is ... and the dataset ... will be chosen and not ..., since it's not relevant for the time.
2. Then, the base models run on the editions and we get the basic segmented and transcribed.
3. Then there are three kinds automatic processes that run on the results and correct the segmentation and transcriptions: (a) an array of rules/scripts that run on the results. E.g. - if all page numbers are on the left side except for one page in the edition - it does not make sense. if the text "PROP. <roman numeral>" appears - it must be proposition header, not paragraph. (b) an array of instructions that give an LLM prompt for how to fix the segmentation and transcription and (c) a set of actual corrections/ground-truth that were already submitted by the researcher. By default the three run one after another and stops, but the researcher can submit overrides, detailing which rules/scripts, LLM instructions and manual corrections run and what is the order. E.g., the researcher can decide that to first we run the manual cirrections, then 3 LLM instuctions, then 2 of the scripts, then another 4 LLM onstructions, then 3 other rules and at the end again the manual corrections. 
4. Once the correction process is done, the researcher can review the results, and edit it - either by adding manual correction, changing the order of the corrections, tweeking iit, adding LLM instruction, etc. The user can keep the layout+tanscription, bringing the flow to a halt, or 
5. This run, creates a new segmetatio and trascription. The researcher can review this, correct the actual results (it'll be added to the actual corrections/ground-truth set), tweek the prompts and re-run the LLMs or confirm. To be economic, we'll cache the result after each LLM prompt runs. Once confirmed, the training re-runs and generates a new model. Each model is stored with id and all the metadata that also includes the step for its generation.
3. Step 2 is repeated how many times we want. Note that a model generated for this edition (or for other edition) can be the input for step #1.

In addition, it's possible to start not from the first step. Let's say we already did a few iterations, we can just use the output of some step and re--run the next, we don't need to do everything all over again. it's important b/c these steps take lots of time.

For ease of use, the flow will run with a dedicated command line interface that supports both get-info stuff and running.
The main principles that mostly automatic research-in-the-loop multi-step, reproducible and that we can start and stop at eavery stage


In terms of development - the approach is that we are not developing a model, but a mostly-automatic procedure that generates models. So, at the end, each edition has a model and the model creation can be reproduced automatically by following the procedure steps. This approach has a few pros. First, it allows us to both have very fitted models. 
In addition, focusing on building an algorithm, and not a model, makes the entire thing reproducible. So if we get more data, we can re-run the algorithm and generate a new model that can be compared to the previous one both manually and automatically.

The output of this process is having two (three?) models per edition: the first segments it, the second transcribes it. 
The input 
To generate this the following steps are done:
per edition and (2) the algo is based on a combination of base models from the communities that are trained in an iterative way with human oversight, LLM that gets prompts and ground truth datasets both from the community and from editions that are "close". This is relevant both for the segmentation and the transcription models.

The first part of this step is to obtain all the relevant datasets available online. 

So the idea is building an interactive software that 

has these steps:
For each edition,
1. The research
1. Runs a base model from scholarship and tune it with ground truth data from datasets available online of "similar" books (e.g., edition from the same century/country/language/typology) + editions that we already transcribed of the Elements 
2. Fixes the results with LLM and scripts with rules -> display the result to the research engineer (me, possibly also Max (?)), the engineer can either (a) correct manually the results or (b) add a prompt so LLM will correct the results -> retrain on corrected pages -> re-run the models.
## High-level Design
Algorithm that builds models, not a single model. Models trained on specific subsets of editions.
What we're actually developing is not the models, but the algorithm, the procedure to build the models.
Using SegmOnto as vocabulary for segmentation labels.
Using Kraken for segmentation and transcription.
Segmentation base: experimenting with different models (U-Net, Mask R-CNN, etc.) - on top of that training, ideally using LLM+scripts to improve results, currently manual correction.
Transcription base: same...
More ground truth data for training: the transcriptions provided by Google Books for some of the editions + available datasets that use the SegmOnto vocabulary + possibly other datasets.
Results stored in eScriptorium instance and exported.

## Current Status
- Infrastructure setup: local eScriptorium instance. We will need to move to a server later, once HumaNum infrastructure is ready.
- Experimenting with different segmentation+transcription models.
- Manual correction of segmentation results.

## Still to be done:
- Finalizing segmentation and transcription algorithm.
- Integrating LLMs to improve segmentation and transcription results (iterative process).
- Integrating additional ground truth data for training.

...
# Publishing
The metadata is available online, in my personal website...