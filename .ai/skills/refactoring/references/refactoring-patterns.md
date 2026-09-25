# Refactoring Patterns

## 1. Extract Method
**Problem**: You have a code fragment that can be grouped together.
**Solution**: Move this code to a separate new method (or function) and replace the old code with a call to the method.

## 2. Inline Method
**Problem**: A method's body is more obvious than the method itself.
**Solution**: Replace calls to the method with the method's content and delete the method itself.

## 3. Extract Variable
**Problem**: You have an expression that's hard to understand.
**Solution**: Place the result of the expression or its parts in separate variables that are self-explanatory.

## 4. Replace Temp with Query
**Problem**: You place the result of an expression in a local variable for later use in your code.
**Solution**: Move the entire expression to a separate method and return the result from it. Query the method instead of using a variable.

## 5. Introduce Parameter Object
**Problem**: Your methods contain a repeating group of parameters.
**Solution**: Replace these parameters with an object.

## 6. Preserve Whole Object
**Problem**: You get several values from an object and then pass them as parameters to a method.
**Solution**: Instead, pass the whole object.

## 7. Replace Magic Number with Symbolic Constant
**Problem**: Your code uses a number that has a certain meaning to it.
**Solution**: Replace this number with a constant that has a human-readable name explaining the meaning of the number.
