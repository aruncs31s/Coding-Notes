---
title: "<% tp.date.now("dddd, MMMM Do, YYYY") %>"
description: "tags::"
tags:
  - ""
cssclasses:
  - wide-page
dg-publish: true
aliases: []
created: 2025-11-04T00:33:55.767992
updated: 2025-11-04T00:33:55.768005
---
tags:: [[+Daily Notes]]

# <% tp.date.now("dddd, MMMM Do, YYYY") %>

← [[<% tp.date.now("YYYY-MM-DD", -1) %>|Yesterday]] | [[<% tp.date.now("YYYY-MM-DD", 1) %>|Tomorrow]] →

---



## 📋 Tasks


### Today's Tasks
```tasks
not done
due on <% tp.date.now("YYYY-MM-DD") %>
```


### Completed Today
```tasks
done on <% tp.date.now("YYYY-MM-DD") %>
```


## 📅 Daily Reflection


### 🙌 What am I excited about today?

- 


### 🎯 What do I want to accomplish today?

- [ ] 
- [ ] 
- [ ] 


### 🤔 What challenges am I facing?

- 


### 🌟 What am I grateful for?

- 


## 📝 Notes

<%* app.workspace.activeLeaf.view.editor?.focus(); %>

---



## 📊 Today's Activity


### Notes Created Today
```dataview
LIST
FROM ""
WHERE dateformat(file.ctime, "yyyy-MM-dd") = "<% tp.date.now("YYYY-MM-DD") %>"
SORT file.ctime ASC
```


### Notes Modified Today
```dataview
LIST
FROM ""
WHERE dateformat(file.mtime, "yyyy-MM-dd") = "<% tp.date.now("YYYY-MM-DD") %>" AND dateformat(file.ctime, "yyyy-MM-dd") != "<% tp.date.now("YYYY-MM-DD") %>"
SORT file.mtime DESC
```

---

*Created: <% tp.date.now("YYYY-MM-DD HH:mm") %>*
