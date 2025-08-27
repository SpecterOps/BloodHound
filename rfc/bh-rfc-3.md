---
bh-rfc: 3
title: Style Guide for Python Coding
authors: |
    [Holms, Alyx](aholms@specterops.io)
    [Rangel, Ulises](urangel@specterops.io)
    [Kaleb, Pomeroy](kpomeroy@specterops.io)
    [Youssef, Kaiboussi](ykaiboussi@specterops.io)

status: DRAFT 
created: 2025-08-25
---

# Style Guide for Python Coding 

## 1. Overview

This Python style guide provides guidelines for writing Python code that is clean, readable, and maintainable. Following conventions improves collaboration and reduces errors.

## 2. Motivation & Goals

-   **Conventions** - Share common understanding when writing python applications @ SpecterOps

# 3. Tooling 
- [Pylint](https://pylint.pycqa.org/en/latest/index.html) code static analyzer to detect coding errors
- [Black](https://black.readthedocs.io/en/stable/the_black_code_style/index.html) formatter adheres to prep8 rules and can be [integrated](https://black.readthedocs.io/en/stable/integrations/editors.html) by most of the IDE's and [VSCode](https://marketplace.visualstudio.com/items?itemName=ms-python.black-formatter)

## 4. Code Layout

#### 4.2 Code Length and Spaces
- **Indentation:** Use four spaces per indentation level. Do not use tabs.
- **Line Length:** Black formatter defaults to 88 characters for code, prep8 recommands 72 characters for comments/docstrings.
- **Blank Lines:** Use blank lines to separate functions and classes, and to separate logical sections within a function.

####  4.3 Import Statements
- Place imports at the top of the file.
- Use separate lines for each import
- Follow this order:
    1. Standard library imports
    2. Related third-party imports
    3. Local application/library specific imports

## 5. Naming Conventions
- **Variable and Function Names:** Use lowercase words separated by underscores (e.g., `my_variable`, `my_function`).
- **Single Variable Names** Avoid using `i`(lower case), `O`(upper case) or `I`(upper case) as single character variable names these letter are indistinguishable on other fonts. Use `L` when templating data at the presentation layer.
- **Class Names:** Use CapitalizedWords convention (e.g., `MyClass`).
- **Constants:** Use all uppercase letters with underscores (e.g., `MAX_LIMIT`).
- **Modules and Packages:** Use short, all-lowercase names. Underscores can be used in module names if it improves readability.

## 6. Comments
- **Block Comments:** Use block comments to explain code logic. Each line should start with a # and a single space.
- **Inline Comments:** Use sparse inline comments to explain a specific part of the code, and ensure a space precedes the comment symbol.

## 7. Docstrings
- Use triple quotes for docstrings at the beginning of modules, classes, and functions.
- Follow the format:

```python  
def function_name(param1: Type, param2: Type) -> ReturnType:  
    """Brief description.      
     Args:  
          param1 (Type): Description.  
          param2 (Type): Description.      
     Returns:  
          ReturnType: Description.  
     """  
```

## 8. Whitespace in Strings, Expressions and Statements
- Avoid extraneous whitespace in expressions and statements.
```python  
# Correct
total = a + b  
```

``` python  
# Incorrect
total =  a   +   b  
```  
- No more than one space around string assigments
```python
# Correct
x = 1
y = 2
long_variable = 3
```

```python
# Incorrect
x             = 1
y             = 2
long_variable = 3
```

## 9. BHCE Recommendations
#### 9.1 Exception Handling

Make use of built-in exception classes when it makes sense. For example, raise a `ValueError` to indicate a programming mistake like a violated precondition, such as may happen when validating function arguments.

```python
if count < 5:  
      raise ValueError(f'Count be at least 5, not {count}.') 
```
#### 9.2 Default Argument Values

Python doesn't support overloaded methods or functions like some other languages do, but you can achieve similar functionality with default arguments. It's a clever way to mimic the behavior of overloading and keep your code clean and readable!

```python
def get_foo(a="default_value"):
..........
```

Beware of mutable objects such as list of dictionary if the function modifies the objects such as appending an item to a list.

```python
# Incorrect
def foo(my_list=[int]):
    my_list.append(20)
    return my_list
    
# Correct
def foo(my_list=None):
    if my_list is None:
      my_list = []
      my_list.append(20)
      return my_list
```