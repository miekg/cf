%{
package cf

import (
	"github.com/miekg/cf/ast"
)

%}

// from: cfengine/core/libpromises/cf3parse.y

// COMMENT will probably be attach to the token and not a token, as that's also not the case in the cf3parse.y
%token IDENTIFIER QSTRING CLASSGUARD PROMISEGUARD BUNDLE BODY PROMISE FATARROW THINARROW NAKEDVAR

// fields inside this union end up as the fields in a structure known
// as ${PREFIX}SymType, of which a reference is passed to the lexer.
%union{
	token ast.Token
}

%%

specification:       /* empty */
                     | blocks
                     {
                        yylex.(*Lexer).Spec = &ast.Specification{}
                        yylex.(*Lexer).Spec.SetChildren([]ast.Node{ast.Up(yylex.(*Lexer).parent)})
                     }

blocks:                block
                     | blocks block;

block:                 bundle
		     {
                        yylex.(*Lexer).yydebug("block:bundle", $$.token)
			// Here we find the actual token, but we get here last. Find original bundle and put
			// token contents in it. Mostly to get the comments out.
			bundle := ast.UpTo(yylex.(*Lexer).parent, &ast.Bundle{})
			if bundle != nil {
				bundle.SetToken($$.token)
			}
		     }
                     | body
                     {
                        yylex.(*Lexer).yydebug("block:body", $$.token)
			body := ast.UpTo(yylex.(*Lexer).parent, &ast.Body{})
			if body != nil {
				body.SetToken($$.token)
			}
                     }
                     | promise
                     | error
                       {
                       }

bundle:                BUNDLE
                       {
                        yylex.(*Lexer).yydebug("bundle:BUNDLE", $$.token)
                        spec := ast.UpTo(yylex.(*Lexer).parent, &ast.Specification{})
                        yylex.(*Lexer).parent = spec
                        b := ast.New(&ast.Bundle{}, $$.token)
                        ast.Append(yylex.(*Lexer).parent, b)
                        yylex.(*Lexer).parent = b
                       }
                       bundletype bundleid arglist bundlebody
                       {
                       }

body:                  BODY
                       {
                        yylex.(*Lexer).yydebug("body:BODY", $$.token)
                        spec := ast.UpTo(yylex.(*Lexer).parent, &ast.Specification{})
                        yylex.(*Lexer).parent = spec
                        b := ast.New(&ast.Body{}, $$.token)
                        ast.Append(yylex.(*Lexer).parent, b)
                        yylex.(*Lexer).parent = b
                       }
                       bodytype bodyid arglist bodybody

promise:               PROMISE
                       {
                            yylex.(*Lexer).yydebug("promise:PROMISE")
                       }
                       promisecomponent promiseid arglist bodybody

bundletype:            bundletype_values

bundletype_values:     typeid
                       {
                       }
                     | error
                       {
                       }

bundleid:              bundleid_values
                       {
                       }

bundleid_values:       symbol
                     | error
                       {
                       }

bodytype:              bodytype_values
                       {
                       }

bodytype_values:       typeid
                       {
                       }
                     | error
                       {
                       }

bodyid:                bodyid_values
                       {
                       }

bodyid_values:         symbol
                     | error
                       {
                       }

promisecomponent:      promisecomponent_values
                       {
                            yylex.(*Lexer).yydebug("promisecomponent")
                       }

promisecomponent_values: typeid
                         {
                         }
                       | error
                         {
                         }

promiseid:             promiseid_values
                       {
                       }

promiseid_values:      symbol
                     | error
                       {
                       }

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

typeid:                IDENTIFIER
                       {
                        ast.Append(yylex.(*Lexer).parent, ast.New(&ast.Identifier{}, $$.token))
                       }

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

symbol:                IDENTIFIER
                       {
                        ast.Append(yylex.(*Lexer).parent, ast.New(&ast.Identifier{}, $$.token))
                       };

arglist:               /* Empty */
                     | arglist_begin aitems arglist_end
                     | arglist_begin arglist_end
                     | arglist_begin error
                       {
                       }

arglist_begin:         '('
                       {
                        yylex.(*Lexer).yydebug("arglist_begin:(", $$.token)
                       }

arglist_end:           ')'
                       {
                        yylex.(*Lexer).yydebug("arglist_end:)", $$.token)
                        bundle := ast.UpTo(yylex.(*Lexer).parent, &ast.Bundle{})
                        if bundle != nil {
                            yylex.(*Lexer).parent = bundle
                        } else { // maybe body?
                            if body := ast.UpTo(yylex.(*Lexer).parent, &ast.Body{}); body != nil {
                                yylex.(*Lexer).parent = body
                            }
                        }
                       }

aitems:                aitem
                       {
		        if _, ok := yylex.(*Lexer).parent.(*ast.ArgList); !ok {
				a := ast.New(&ast.ArgList{})
				ast.Append(yylex.(*Lexer).parent, a)
				yylex.(*Lexer).parent = a
		        }
                        al := ast.New(&ast.ArgListItem{}, $$.token)
                        ast.Append(yylex.(*Lexer).parent, al)
                       }
                     | aitem ','
                     | aitem ',' aitems
                       {
		        if _, ok := yylex.(*Lexer).parent.(*ast.ArgList); !ok {
				a := ast.New(&ast.ArgList{})
				ast.Append(yylex.(*Lexer).parent, a)
				yylex.(*Lexer).parent = a
                        }
                        al := ast.New(&ast.ArgListItem{}, $$.token)
                        ast.Append(yylex.(*Lexer).parent, al)
		       }

aitem:                 IDENTIFIER  /* recipient of argument is never a Literal */
                       {
                       }
                     | error
                       {
                       }

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

bundlebody:            body_begin
                       {
                       }

                       bundle_decl

                       '}'
                       {
                        // only here for comments.
                        if bundle := ast.UpTo(yylex.(*Lexer).parent, &ast.Bundle{}); bundle != nil {
                            bundle.SetToken($$.token)
                        }
                       }

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

body_begin:            '{'
                       {
                       }
                     | error
                       {
                       }


/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

bundle_decl:           /* empty */
                     | bundle_statements

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

bundle_statements:     bundle_statement
                     | bundle_statements bundle_statement
                     | error
                       {
                       }


bundle_statement:      promise_guard
                       {
                       }
                       classpromises_decl
                       {
                       }

promise_guard:         PROMISEGUARD             /* BUNDLE ONLY */
                       {
                        yylex.(*Lexer).yydebug("promise_guard", $$.token)
                        pg := ast.New(&ast.PromiseGuard{}, $$.token)
                        // If there is previous promiseguard, this one closes it, and we can reyylex.(*Lexer).parent this new one, to that _parent_.
                        prev := ast.UpTo(yylex.(*Lexer).parent, &ast.PromiseGuard{})
                        if prev != nil {
                            yylex.(*Lexer).parent = prev.Parent()
                        }
                        ast.Append(yylex.(*Lexer).parent, pg)
                        yylex.(*Lexer).parent = pg
                       }

classpromises_decl:    /* empty */
                       {
                       }
                     | classpromises
                       {
                       }

classpromises:         classpromise
                       {
                       }
                     | classpromises
                       {
                       }
                       classpromise
                       {
                       }

classpromise:          class
                       {
                            yylex.(*Lexer).yydebug("classpromise", $$.token)
                       }
                     | promise_decl
                       {
                       }

promise_decl:          promise_line ';'
                       {
                            yylex.(*Lexer).yydebug("promise_decl", $$.token)
                       }
                     | promiser error
                       {
                       }

promise_line:          promise_with_promisee
                       {
                       }
                     | promise_without_promisee
                      {
                      }


promise_with_promisee: promiser

                       promisee_arrow

                       rval
                       {
                       }

                       promise_decl_constraints

promise_without_promisee: promiser
                       {
                        yylex.(*Lexer).yydebug("promise_without_promisee: promiser", $$.token)
                       }

                       promise_decl_constraints
                       {
                        yylex.(*Lexer).yydebug("promise_without_promisee: promise_decl_constraints", $$.token)
                       }

promiser:              QSTRING
                       {
                        yylex.(*Lexer).yydebug("promiser:QSTRING", $$.token)
                        // same level as previous Promiser, or PromiseGuard

                        prev := ast.UpTo(yylex.(*Lexer).parent, &ast.Promiser{})
                        if prev == nil {
                            if prev = ast.UpTo(yylex.(*Lexer).parent, &ast.PromiseGuard{}); prev != nil {
                                yylex.(*Lexer).parent = prev
                            }
                        } else {
                            yylex.(*Lexer).parent = prev.Parent()
                        }

                        p := ast.New(&ast.Promiser{}, $$.token)
                        ast.Append(yylex.(*Lexer).parent, p)
                        yylex.(*Lexer).parent = p
                       }

promise_decl_constraints:       /* empty */
                              | constraints_decl
                              | constraints_decl error
                                {
                                }

constraints_decl:      constraints
                       {
                       }

constraints:           constraint                           /* BUNDLE ONLY */
                     | constraints ',' constraint

constraint:            constraint_id                        /* BUNDLE ONLY */
                       {
                       }
                       assign_arrow
                       {
                       }
                       rval
                       {
                        yylex.(*Lexer).yydebug("contraint:rval")
                       }

constraint_id:         IDENTIFIER                        /* BUNDLE ONLY */
                       {
                        yylex.(*Lexer).yydebug("contraint_id:IDENTIFIER", $$.token)

                        prev := ast.UpTo(yylex.(*Lexer).parent, &ast.Promiser{})
                        if prev != nil {
                            yylex.(*Lexer).parent = prev
                        }

                        c := ast.New(&ast.Constraint{}, $$.token)
                        ast.Append(yylex.(*Lexer).parent, c)
                        yylex.(*Lexer).parent = c
                       }
                     | error
                       {
                       }

bodybody:              body_begin
                       {
                       }

                       bodybody_inner

                       '}'
                       {
                        // only here for comments.
                        if body := ast.UpTo(yylex.(*Lexer).parent, &ast.Body{}); body != nil {
                            body.SetToken($$.token)
                        }
                       }

bodybody_inner:        /* empty */
                     | bodyattribs

bodyattribs:           bodyattrib                    /* BODY/PROMISE ONLY */
                     | bodyattribs bodyattrib

bodyattrib:            class
                       {
                       }
                     | selection_line

selection_line:        selection ';'
                     | selection error
                       {
                       }

selection:             selection_id                         /* BODY/PROMISE ONLY */
                       {
                        yylex.(*Lexer).yydebug("selection:selection_id", $$.token)
                       }
                       assign_arrow
                       {
                        yylex.(*Lexer).yydebug("selection:assign_arrow")
                       }
                       rval
                       {
                        yylex.(*Lexer).yydebug("selection:rval", $$.token)
                       }

selection_id:          IDENTIFIER
                       {
                        yylex.(*Lexer).yydebug("selection_id:IDENTIFIER", $$.token)
                        // need to be parent of body.
			body := ast.UpTo(yylex.(*Lexer).parent, &ast.Body{})
                        yylex.(*Lexer).parent = body

                        s := ast.New(&ast.Selection{}, $$.token)
                        ast.Append(yylex.(*Lexer).parent, s)
                        yylex.(*Lexer).parent = s
                       }
                     | error
                       {
                       }

assign_arrow:          FATARROW
                       {
                        ast.Append(yylex.(*Lexer).parent, ast.New(&ast.FatArrow{}, $$.token))
                       }
                     | error
                       {
                       }

promisee_arrow:        THINARROW
                       {
                       }

class:                 CLASSGUARD
                       {
                        yylex.(*Lexer).yydebug("class")
                        gc := ast.New(&ast.ClassGuard{}, $$.token)
                        // If there is previous classguard, this one closes it, and we can reyylex.(*Lexer).parent this new one, to that _parent_.
                        prev := ast.UpTo(yylex.(*Lexer).parent, &ast.ClassGuard{})
                        // If there is no previous one, look for promise guard, and yylex.(*Lexer).parent to that.
                        if prev == nil {
                            prev = ast.UpTo(yylex.(*Lexer).parent, &ast.PromiseGuard{})
                        }
                        // re-yylex.(*Lexer).parent if found
                        if prev != nil {
                            yylex.(*Lexer).parent = prev.Parent()
                        }

                        ast.Append(yylex.(*Lexer).parent, gc)
                        yylex.(*Lexer).parent = gc
                       }

rval:                  IDENTIFIER
                       {
                        // awkward that these are the only ones here..?
                        yylex.(*Lexer).yydebug("rval:IDENTIFIER", $$.token)
                        i := ast.New(&ast.Identifier{}, $$.token)
                        ast.Append(yylex.(*Lexer).parent, i)
                       }
                     | QSTRING
                       {
                        yylex.(*Lexer).yydebug("rval:QSTRING", $$.token)
                        q := ast.New(&ast.Qstring{}, $$.token)
                        ast.Append(yylex.(*Lexer).parent, q)
                       }
                     | NAKEDVAR
                       {
                       }
                     | list
                       {
                       }
                     | usefunction
                       {
                       }
                     | error
                       {
                       }

list:                  '{' '}'
                     | '{' Litems '}'
                     | '{' Litems ',' '}'

Litems:
                       Litem
		       {
		        // add yylex.(*Lexer).parent list
		        if _, ok := yylex.(*Lexer).parent.(*ast.List); !ok {
				l := ast.New(&ast.List{})
				ast.Append(yylex.(*Lexer).parent, l)
				yylex.(*Lexer).parent = l
			}
                        l := ast.New(&ast.ListItem{}, $$.token)
                        ast.Append(yylex.(*Lexer).parent, l)
		       }
                     | Litems ',' Litem
		       {
		        // add yylex.(*Lexer).parent list
		        if _, ok := yylex.(*Lexer).parent.(*ast.List); !ok {
				l := ast.New(&ast.List{})
				ast.Append(yylex.(*Lexer).parent, l)
				yylex.(*Lexer).parent = l
			}
                        l := ast.New(&ast.ListItem{}, $3.token)
                        ast.Append(yylex.(*Lexer).parent, l)
		       }
                     | Litem error
                       {
                       }

Litem:                 IDENTIFIER
                       {
                       }
                     | QSTRING
                       {
                       }
                     | NAKEDVAR
                       {
                       }
                     | usefunction
                       {
                       }
                     | error
                       {
                       }

functionid:            IDENTIFIER
                       {
                        f := ast.New(&ast.Function{}, ast.Token{})
                        ast.Append(yylex.(*Lexer).parent, f)
                        yylex.(*Lexer).parent = f

                        ast.Append(yylex.(*Lexer).parent, ast.New(&ast.Identifier{}, $$.token))
                       }
                     | NAKEDVAR
                       {
                        f := ast.New(&ast.Function{}, ast.Token{})
                        ast.Append(yylex.(*Lexer).parent, f)
                        yylex.(*Lexer).parent = f

                        ast.Append(yylex.(*Lexer).parent, ast.New(&ast.NakedVar{}, $$.token))
                       }

usefunction:           functionid
                       {
                       }
                       givearglist

givearglist:           '('
                       {
                       }

                       gaitems
                       {
                       }

                       ')'
                       {
                       }

gaitems:               /* empty */
                     | gaitem
                       {
                        l:= ast.New(&ast.GiveArgItem{}, $$.token) // single arg
                        ast.Append(yylex.(*Lexer).parent, l)
                       }
                     | gaitems ',' gaitem
                       {
                        l:= ast.New(&ast.GiveArgItem{}, $3.token) // multiple args
                        ast.Append(yylex.(*Lexer).parent, l)
                       }
                     | gaitem error
                       {
                       }

gaitem:                IDENTIFIER
                     | QSTRING
                     | NAKEDVAR
                     | usefunction
                     | error
%%
