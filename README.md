# `voltctl` - A command line tools to access VOLTHA
In today's VOLTHA implementation the operator interacts with VOLTHA via a CLI
that is accessed by `SSH`ing to the VOLTHA tool. While is can be convenent as 
it requires not external tool to be installed it is an abirtation in terms of 
industry trends for tools such as `docker` and `kubernetes`, which are both
tools that VOLTHA leverages.

This repository contains a tool that attempts to provide a use model for
VOLTHA that is similar to that of `docker` and `kubernetes` in that a simple
control application is provided to invoke various funcs and the output can 
be displayed as a customized/filtered table or as `JSON`.

WIP - just barely started

