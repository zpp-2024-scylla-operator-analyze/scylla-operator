# Operator Analyze Tool

## Summary

This proposal aims to introduce new `scylla-operator analyze` command line tool for automatic detection of erroneous
conditions, specified as predefined rules, in Kubernetes clusters. The tool is going to support static analysis based
on `must-gather` archives or live clusters. 

## Motivation

ScyllaDB receives many recurring issue reports. By introducing this tool, we could
optimize the troubleshooting process.

### Goals

List the specific goals of the proposal. How will we know that this has succeeded?
- Outline the design of `analyze` tool.
- 

### Non-Goals

- Monitoring of live clusters
- Automatic resolution of errors

## -Proposal

This is where we get down to the specifics of what the enhancement actually is.
This should have enough detail that reviewers can understand exactly what
you're proposing, but should not include things like API designs or
implementation. What is the desired outcome and how do we measure success?.
The "Design Details" section below is for the real nitty-gritty.

### -User Stories

Detail the things that people will be able to do if this is implemented.
Include as much detail as possible so that people can understand the "how" of
the system. The goal here is to make this feel real for users without getting
bogged down.

#### -Story 1

#### -Story 2

### -Notes/Constraints/Caveats [Optional]

What are the caveats to the proposal?
What are some important details that didn't come across above?
Go in to as much detail as necessary here.
This might be a good place to talk about core concepts and how they relate.

### -Risks and Mitigations

What are the risks of this proposal, and how do we mitigate? Think broadly.
For example, consider both security and how this will impact the ecosystem.

## -Design Details

This section should contain enough information that the specifics of your
change are understandable. This may include API specs (though not always
required) or even code snippets. If there's any ambiguity about HOW your
proposal will be implemented, this is the place to discuss them.

### -Test Plan

Consider the following in developing a test plan for this enhancement:
- Will there be e2e tests, in addition to unit tests?

No need to outline all of the test cases, just the general strategy. Anything
that would count as tricky in the implementation, and anything particularly
challenging to test, should be called out.

All code is expected to have adequate tests.

### Upgrade / Downgrade Strategy

No specific action for upgrades / downgrades is needed.

### Version Skew Strategy

The tool is mostly self-isolated, there should be no tricky interactions.

## -Implementation History

Major milestones in the life cycle of a proposal should be tracked in `Implementation
History`.

## Drawbacks

Due to the dynamic nature of software involved, the rules might need to be updated frequently,
increasing maintenance efforts to keep a satisfactory diagnosis accuracy.
    
## -Alternatives

What other approaches did you consider, and why did you rule them out? These do
not need to be as detailed as the proposal, but should include enough
information to express the idea and why it was not acceptable.
