## Three algorithms for getting result row numbers via JsonPath

Article Link:

## Introduction 

First we have to figure out what problem the algorithm solves?
The problem arises from a business scenario: when processing json we need to use a path similar to the jsonpath to get the line number of a specific object in the target json (note that this is not the result of obtaining the object, but rather to get the line number of the first object that matches , i.e., the offset relative to the line number of the starting line of the json)


### for example
Here is a json, I want to know the line number where user_1.user_detail.links.self matches the result i.e. 16, what should I do?
```
{
  "user_1": {
    "name": "darkFernMoss",
    "hobbies": [],
    "user_detail": {
      "pwd_strength": "Strong",
      "create_time": "2023-04-24 02:57:17.0",
      "settings": [
        "dark",
        "female",
        "ice-cream"
      ],
      "links": {
        "next": "",
        "previous": "",
        "self": "https://iam.cn-east-3.myhuaweicloud.com"
      },
      "isDeleted": false
    }
  }
}
```


### Algorithmic Premise and Details

1.  The starting line number of json starts at 1.
2.  If the target of the match is an array, then the array can only be the final value, and continued addressing from an element of the array is not supported.
3.  When calculating the line number may **pass by** a variety of json objects, which need to be cumulative of their line numbers. Such as by the simplest single line (user_1.name), or by nested json (user_1.user_detail), or by a json array (user_1.user_detail.settings), and it is still possible that each object in the array is composed of all three.
4.  When accumulating json of array type, if the array is empty and the "[]" brackets are not open, the line number it occupies will still be 1, e.g. (user_1.hobbies).

## Solution 1 (mark-recapture method)

Idea: in the first traversal will be matched to the (key-value) pairs of key tagged (here can be understood as the key to mark, such as adding a fixed suffix, here I add the suffix is _cspm_highlight, easy to find the line again), in the tagged, the use of json.indent will be json data formatting, the purpose is to be able to add line breaks between each line (the default json.marshal is the most compact formatting, does not contain line breaks), and then through a traversal of the cumulative number of line breaks, while the record has been tagged with the key where the line number.

## Solution 2 (recursive solution before optimization)

The last solution is easy to think of but the code is longer, the details are complicated, in order to count the line number to reformat the json data, then can not count the number of line breaks through the number of direct access to the line number?Idea: in the recursion if you can not match the next object in the jsonpath on the return of false and line number, and the return of the line number will not be counted in the answer, if the match to the next object, then through another recursive function to add the line number of the current road has been passed. If you have reached the last jsonpath and the value also matches, then return true and the current line number +1.


## Solution 3 (optimized version of recursive solution)

Idea: This time we traverse the data json down in recursion, while maintaining a string path to record the jsonpath that has been passed, if this jsonpath exists in the alarmJson, then the result will be added to the answer.The biggest optimization point of this recursive scheme is that we use reverse thinking to match a given map composed of alarmJson by maintaining the path while traversing the dataJson, making full use of the already computed row numbers without double computation.
