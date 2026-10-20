---
title: Railroad
description: The rules of a grammar drawn as tracks, which a valid input follows from start to end.
beta: true
featured: 2
---

## A JSON number

```mermaid
railroad-ebnf-beta
title "JSON number"

number = [ "-" ] integer [ fraction ] [ exponent ] ;
integer = "0" | onenine { digit } ;
fraction = "." digit { digit } ;
exponent = ( "e" | "E" ) [ "+" | "-" ] digit { digit } ;
onenine = ? a digit from 1 to 9 ? ;
digit = "0" | onenine ;
```

## A small expression grammar

```mermaid
railroad-ebnf-beta
title "Arithmetic with calls"

expr ::= term (("+" | "-") term)* ;
term ::= factor (("*" | "/" | "%") factor)* ;
factor ::= "-"? (number | call | name | "(" expr ")") ;
call ::= name "(" (expr ("," expr)*)? ")" ;
```

## A semantic version in ABNF

```mermaid
railroad-abnf-beta
title "Semantic version"

version = core [ "-" pre-release ] [ "+" build ] ;
core = number "." number "." number ;
pre-release = identifier *( "." identifier ) ;
build = identifier *( "." identifier ) ;
number = "0" / ( %x31-39 *DIGIT ) ;
identifier = 1*( ALPHA / DIGIT / "-" ) ;
```

## An INI file line in PEG

```mermaid
railroad-peg-beta
title "INI file line"

Line <- Section / Comment / Pair ;
Section <- "[" Name "]" ;
Comment <- ";" (!EndOfLine .)* ;
Pair <- Name "=" Value ;
Name <- Letter (Letter / Digit / "_")* ;
Value <- (!EndOfLine .)* ;
```

## A Go duration, written as calls

```mermaid
railroad-beta
title "Go duration"

duration = sequence(
    optional(choice(terminal("+"), terminal("-"))),
    oneOrMore(sequence(nonterminal("amount"), nonterminal("unit")))
) ;
amount = sequence(
    oneOrMore(special("digit")),
    optional(sequence(terminal("."), oneOrMore(special("digit"))))
) ;
unit = choice(terminal("h"), terminal("m"), terminal("s"), terminal("ms"), terminal("us"), terminal("ns")) ;
```
