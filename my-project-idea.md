# ZAP 
ZAP is an AI-powered **CLI/TUI** tool that can be used to generate API requests from natural language prompts.
Think of it as **claude code** with the functionalities of **postman**  or  think of **httpie** with **AI powered features**.I want zap to be a tool that understands your codebase, it able to test your API endpoints through natural language from the user and also fix code like claude code( HUMAN IN LOOP APPROVAL)


# How other tools are able to get the knowledge of the codebase
The **Cursor architecture** functions as a high-speed, local search engine, prioritizing scale and retrieval speed by pre-indexing the entire codebase into a vector database. It uses a background daemon to monitor file system events, parses code into logical structural blocks (functions/classes) using AST analysis (tree-sitter), and stores semantic embeddings of these chunks. This allows the system to handle massive projects (50,000+ files) by instantly retrieving relevant context via cosine similarity search, though it relies on complex "Merkle-style" synchronization to ensure the stored vectors never drift from the actual file content.

In contrast, the **Claude Code architecture** operates as an autonomous agent, treating the codebase as a live environment to be explored dynamically rather than a dataset to be indexed. Instead of querying a pre-built database, it utilizes a large context window and "tool use" capabilities to execute real-time shell commands—such as `ls` to map directory structure, `grep` to locate patterns, and `cat` to read files—based on the specific user request. This approach ensures the model always works with the absolute "ground truth" of the code and mimics human debugging workflows, but it trades the millisecond-latency of vector retrieval for the slower, iterative process of active discovery.

**Currently , i dont know the best one for the application**

## Tech stack 
Go-lang 
bubble tea 
ligloss
bubbles
and other libraries from charm ecosystem that can be used in this application. 
ollama ( LLM for development stage)


Also I want think of building an extension for this tool to be used in vs code, cursor and antigravity.I wanted something similar to the claude code extension in vscode ( it opens as a side panel in vscode and we can interact with it from there) and also we can use it in the terminal.

## Knowledge and memory of the agent.
With some AI coding agents that are coming up, they have a way of storing the progress and knowledge of what they were doing locally. so I want the tool to be like that. I want it to be able to remember the previous conversations and the context of the project. So i was thinking of creating a .zap folder, inside the project folder to store the knowledge and memory of the agent. with this, if  i were to create an extension , the terminal and the extension can share the same knowledge and memory.



let us analyze this project and start building it ( MVP first). 