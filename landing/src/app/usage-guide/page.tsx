export default function GettingStarted() {
  return (
    <div className="min-h-screen bg-white text-stone-700">
      {/* Fixed Navigation */}
      <nav className="fixed top-0 left-0 right-0 z-50 p-2 text-sm bg-white flex items-center justify-between border-b border-stone-200">
        <a href="/" className="h-[2em] px-2 hover:bg-stone-200 hover:rounded-md inline-flex items-center">
          <div className="w-4 h-4 logo-gradient rounded-full mr-2 flex items-center justify-center">
            <span className="text-white font-bold text-[11px] logo-text">R</span>
          </div>
          Rishi
        </a>
        <div className="flex items-center gap-1">
          <a
            href="/getting-started"
            className="h-[2em] px-2 hover:bg-stone-200 hover:rounded-md flex items-center"
          >
            <svg
              width="18"
              height="18"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
              className="mr-1"
              aria-hidden="true"
            >
              <path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z" />
              <polyline points="14 2 14 8 20 8" />
              <line x1="16" y1="13" x2="8" y2="13" />
              <line x1="16" y1="17" x2="8" y2="17" />
              <line x1="10" y1="9" x2="8" y2="9" />
            </svg>
            Usage Guide
          </a>
          <a
            href="https://github.com/Omkar-Waingankar/rishi"
            className="h-[2em] px-2 hover:bg-stone-200 hover:rounded-md flex items-center"
            target="_blank"
            rel="noopener noreferrer"
            aria-label="GitHub"
          >
            <svg
              width="18"
              height="18"
              viewBox="0 0 98 96"
              fill="currentColor"
              className="mr-1"
              aria-hidden="true"
            >
              <path
                fillRule="evenodd"
                clipRule="evenodd"
                d="M48.854 0C21.839 0 0 22 0 49.217c0 21.756 13.993 40.172 33.405 46.69 2.427.49 3.316-1.059 3.316-2.362 0-1.141-.08-5.052-.08-9.127-13.59 2.934-16.42-5.867-16.42-5.867-2.184-5.704-5.42-7.17-5.42-7.17-4.448-3.015.324-3.015.324-3.015 4.934.326 7.523 5.052 7.523 5.052 4.367 7.496 11.404 5.378 14.235 4.074.404-3.178 1.699-5.378 3.074-6.6-10.839-1.141-22.243-5.378-22.243-24.283 0-5.378 1.94-9.778 5.014-13.2-.485-1.222-2.184-6.275.486-13.038 0 0 4.125-1.304 13.426 5.052a47 47 0 0 1 12.214-1.63c4.125 0 8.33.571 12.213 1.63 9.302-6.356 13.427-5.052 13.427-5.052 2.67 6.763.97 11.816.485 13.038 3.155 3.422 5.015 7.822 5.015 13.2 0 18.905-11.404 23.06-22.324 24.283 1.78 1.548 3.316 4.481 3.316 9.126 0 6.6-.08 11.897-.08 13.526 0 1.304.89 2.853 3.316 2.364 19.412-6.52 33.405-24.935 33.405-46.691C97.707 22 75.788 0 48.854 0"
              />
            </svg>
            Github
          </a>
        </div>
      </nav>

      {/* Main Content */}
      <main className="my-32 mx-auto max-w-2xl text-md pb-16 px-4">
        <article>
          <h1 className="text-4xl font-bold mb-8">Rishi Usage Guide</h1>

          <p className="mb-6">
            Rishi is an AI-powered coding agent designed specifically for R users. This comprehensive guide will help you make the most of Rishi's capabilities and integrate it effectively into your workflow.
          </p>

          {/* Table of Contents */}
          <div className="bg-stone-50 rounded-md p-6 mb-10 border border-stone-200">
            <h1 className="text-2xl font-semibold mb-4">Table of Contents</h1>
            <nav className="space-y-2">
              <div>
                <a href="#installation" className="text-blue-600 hover:text-blue-800 hover:underline">Installation</a>
              </div>
              <div>
                <a href="#core-features" className="text-blue-600 hover:text-blue-800 hover:underline">Core Features</a>
                <div className="ml-4 mt-1 space-y-1 text-sm">
                  <div><a href="#available-models" className="text-blue-600 hover:text-blue-800 hover:underline">Available Models</a></div>
                  <div><a href="#context-integration" className="text-blue-600 hover:text-blue-800 hover:underline">Context Integration</a></div>
                  <div><a href="#available-tools" className="text-blue-600 hover:text-blue-800 hover:underline">Available Tools</a></div>
                </div>
              </div>
              <div>
                <a href="#use-cases" className="text-blue-600 hover:text-blue-800 hover:underline">Common Use Cases</a>
                <div className="ml-4 mt-1 space-y-1 text-sm">
                  <div><a href="#eda" className="text-blue-600 hover:text-blue-800 hover:underline">Exploratory Data Analysis</a></div>
                  <div><a href="#debugging" className="text-blue-600 hover:text-blue-800 hover:underline">Debugging Console Errors</a></div>
                  <div><a href="#visualization" className="text-blue-600 hover:text-blue-800 hover:underline">Visualization Refinement</a></div>
                  <div><a href="#learning" className="text-blue-600 hover:text-blue-800 hover:underline">Learning New Packages</a></div>
                  <div><a href="#refactoring" className="text-blue-600 hover:text-blue-800 hover:underline">Code Refactoring</a></div>
                </div>
              </div>
              <div>
                <a href="#best-practices" className="text-blue-600 hover:text-blue-800 hover:underline">Best Practices</a>
              </div>
              <div>
                <a href="#effective-prompts" className="text-blue-600 hover:text-blue-800 hover:underline">Tips for Effective Prompts</a>
              </div>
              <div>
                <a href="#getting-help" className="text-blue-600 hover:text-blue-800 hover:underline">Getting Help</a>
              </div>
              <div>
                <a href="#privacy" className="text-blue-600 hover:text-blue-800 hover:underline">Privacy and Security</a>
              </div>
            </nav>
          </div>

          {/* Installation */}
          <h2 id="installation" className="text-2xl font-semibold mt-10 mb-4">Installation</h2>
          <p className="mb-4">
            Install Rishi from GitHub and launch it as a RStudio Addin:
          </p>
          <div className="bg-stone-50 rounded-md p-4 font-mono text-sm mb-4 overflow-x-auto">
            <code>
              <span className="text-stone-600"># Install from GitHub</span><br />
              remotes::install_github("Omkar-Waingankar/rishi", subdir = "addin")<br />
              <br />
              <span className="text-stone-600"># Launch the Addin</span><br />
              rishi:::rishiAddin()
            </code>
          </div>
          <p className="mb-4">
            You'll need an API key from either <a href="https://console.anthropic.com/api-keys" className="text-blue-600 hover:text-blue-800 underline">Anthropic</a> or <a href="https://platform.openai.com/api-keys" className="text-blue-600 hover:text-blue-800 underline">OpenAI</a>. Configure your keys in Rishi's settings when you first launch it.
          </p>

          {/* Core Features */}
          <h2 id="core-features" className="text-2xl font-semibold mt-10 mb-4">Core Features</h2>

          <h3 id="available-models" className="text-xl font-semibold mt-6 mb-3">Available Models</h3>
          <p className="mb-4">
            Rishi supports multiple AI providers:
          </p>
          <ul className="list-disc pl-6 mb-4 space-y-2">
            <li><strong>Anthropic:</strong> Claude 4.5 Sonnet, Claude 4 Sonnet, Claude 3.7 Sonnet</li>
            <li><strong>OpenAI:</strong> GPT-5, GPT-4o, GPT-4o-mini</li>
          </ul>
          <p className="mb-4">
            Switch between models using the dropdown in the chat interface based on your task requirements and API availability.
          </p>

          <h3 id="context-integration" className="text-xl font-semibold mt-6 mb-3">Context Integration</h3>
          <p className="mb-4">
            Unlike generic AI chatbots, Rishi understands your RStudio environment. You can provide context in several ways:
          </p>
          <ul className="list-disc pl-6 mb-4 space-y-2">
            <li><strong>Active Tab:</strong> Include the currently open file in RStudio to give Rishi direct visibility into the code you're working on</li>
            <li><strong>Current Plot:</strong> Share your active plot from the Plots pane for feedback, improvements, or explanations</li>
            <li><strong>Images:</strong> Upload screenshots, diagrams, or data visualizations (up to 5MB each, max 3 images) to provide visual context</li>
          </ul>
          <p className="mb-4">
            Access these options through the context dropdown in the input area at the bottom of the chat interface.
          </p>

          <h3 id="available-tools" className="text-xl font-semibold mt-6 mb-3">Available Tools</h3>
          <p className="mb-4">
            Rishi can interact with your R environment through several tools:
          </p>
          <ul className="list-disc pl-6 mb-4 space-y-2">
            <li><strong>File Operations:</strong> List, read, create, and edit files in your working directory</li>
            <li><strong>Console Execution:</strong> Run R code directly in your RStudio console</li>
            <li><strong>R Help:</strong> Access R documentation and help files to guarantee code correctness</li>
          </ul>
          <p className="mb-4">
            Rishi automatically selects the appropriate tools based on your requests. You don't need to manually invoke them.
          </p>

          {/* Use Cases */}
          <h2 id="use-cases" className="text-2xl font-semibold mt-10 mb-4">Common Use Cases</h2>

          <h3 id="eda" className="text-xl font-semibold mt-6 mb-3">Exploratory Data Analysis</h3>
          <p className="mb-4">
            Ask Rishi to help you explore a new dataset:
          </p>
          <div className="bg-stone-50 rounded-md p-4 mb-4 italic">
            "I just loaded a dataset called customer_data. Can you help me understand its structure and identify any interesting patterns?"
          </div>
          <p className="mb-4">
            Rishi can examine your data, generate summary statistics, create visualizations, and suggest next steps for analysis.
          </p>

          <h3 id="debugging" className="text-xl font-semibold mt-6 mb-3">Debugging Console Errors</h3>
          <p className="mb-4">
            When you encounter cryptic error messages:
          </p>
          <div className="bg-stone-50 rounded-md p-4 mb-4 italic">
            "I'm getting an error 'object of type 'closure' is not subsettable' when running my analysis. Can you help me figure out what's wrong?"
          </div>
          <p className="mb-4">
            Rishi can identify the problematic code, explain the error, and suggest fixes.
          </p>

          <h3 id="visualization" className="text-xl font-semibold mt-6 mb-3">Visualization Refinement</h3>
          <p className="mb-4">
            Get help polishing your plots:
          </p>
          <div className="bg-stone-50 rounded-md p-4 mb-4 italic">
            "Can you help me improve this plot? I want better colors, clearer labels, and a more professional look."
          </div>
          <p className="mb-4">
            Use the "Current Plot" context option to share your visualization, and Rishi will suggest improvements with ready-to-run ggplot2 code.
          </p>

          <h3 id="learning" className="text-xl font-semibold mt-6 mb-3">Learning New Packages</h3>
          <p className="mb-4">
            Ramp up quickly on unfamiliar packages:
          </p>
          <div className="bg-stone-50 rounded-md p-4 mb-4 italic">
            "I need to use the tidymodels package for the first time. Can you show me how to build a simple linear regression model?"
          </div>
          <p className="mb-4">
            Rishi can dive into the package documentation and provide specific guidance with code examples tailored to your use case.
          </p>

          <h3 id="refactoring" className="text-xl font-semibold mt-6 mb-3">Code Refactoring</h3>
          <p className="mb-4">
            Improve existing code:
          </p>
          <div className="bg-stone-50 rounded-md p-4 mb-4 italic">
            "This function works but it's slow and hard to read. Can you help me refactor it to be more efficient and maintainable?"
          </div>
          <p className="mb-4">
            Include your active tab as context so Rishi can see the full code and provide specific refactoring suggestions.
          </p>

          {/* Best Practices */}
          <h2 id="best-practices" className="text-2xl font-semibold mt-10 mb-4">Best Practices</h2>

          <h3 className="text-xl font-semibold mt-6 mb-3">Use Version Control (Git)</h3>
          <p className="mb-4">
            <strong className="text-red-600">Important:</strong> Always use Git or another version control system with Rishi. This protects you from accidental changes:
          </p>
          <ul className="list-disc pl-6 mb-4 space-y-2">
            <li>Rishi can create and modify files in your working directory</li>
            <li>While Rishi is generally careful, mistakes can happen</li>
            <li>Git allows you to easily review changes and revert if needed</li>
            <li>Commit your work frequently when using AI assistance</li>
          </ul>
          <div className="bg-stone-50 rounded-md p-4 font-mono text-sm mb-4 overflow-x-auto">
            <code>
              <span className="text-stone-600"># Initialize git if you haven't already</span><br />
              git init<br />
              git add .<br />
              git commit -m "Initial commit before using Rishi"
            </code>
          </div>

          <h3 className="text-xl font-semibold mt-6 mb-3">Set a Working Directory</h3>
          <p className="mb-4">
            Rishi requires a working directory to be set in RStudio. This safety feature prevents Rishi from accessing arbitrary locations on your computer. Either:
          </p>
          <ul className="list-disc pl-6 mb-4 space-y-2">
            <li>Open an R project (.Rproj file), or</li>
            <li>Use <code className="bg-stone-100 px-1 py-0.5 rounded">setwd()</code> to set your working directory explicitly</li>
          </ul>

          <h3 className="text-xl font-semibold mt-6 mb-3">Provide Clear Context</h3>
          <p className="mb-4">
            The more context you provide, the better Rishi's responses:
          </p>
          <ul className="list-disc pl-6 mb-4 space-y-2">
            <li>Describe what you're trying to accomplish, not just what you want the code to do</li>
            <li>Include relevant active tabs when asking about specific code</li>
            <li>Share plots when asking for visualization improvements</li>
            <li>Mention any constraints or preferences (e.g., "using tidyverse style")</li>
          </ul>

          <h3 className="text-xl font-semibold mt-6 mb-3">Manage Your API Keys</h3>
          <p className="mb-4">
            Rishi uses your own API keys, giving you control over costs and usage:
          </p>
          <ul className="list-disc pl-6 mb-4 space-y-2">
            <li>Monitor your API usage through your provider's dashboard</li>
            <li>Set spending limits on your API accounts to avoid surprises</li>
            <li>Use less expensive models (GPT-4o-mini, Claude 3.7 Sonnet) for simpler tasks</li>
            <li>Your keys are stored locally and never sent to external servers (except the AI providers)</li>
          </ul>

          {/* Tips for Effective Prompts */}
          <h2 id="effective-prompts" className="text-2xl font-semibold mt-10 mb-4">Tips for Effective Prompts</h2>

          <h3 className="text-xl font-semibold mt-6 mb-3">Be Specific</h3>
          <div className="bg-red-50 border-l-4 border-red-400 p-4 mb-2">
            <p className="font-semibold text-red-800">Less effective:</p>
            <p className="text-red-700 italic">"Make a plot"</p>
          </div>
          <div className="bg-green-50 border-l-4 border-green-400 p-4 mb-4">
            <p className="font-semibold text-green-800">More effective:</p>
            <p className="text-green-700 italic">"Create a scatter plot of price vs. mileage from the cars dataset, colored by manufacturer, with a smoothed trend line"</p>
          </div>

          <h3 className="text-xl font-semibold mt-6 mb-3">Mention Your Preferences</h3>
          <div className="bg-red-50 border-l-4 border-red-400 p-4 mb-2">
            <p className="font-semibold text-red-800">Less effective:</p>
            <p className="text-red-700 italic">"Clean this data"</p>
          </div>
          <div className="bg-green-50 border-l-4 border-green-400 p-4 mb-4">
            <p className="font-semibold text-green-800">More effective:</p>
            <p className="text-green-700 italic">"Clean this data using tidyverse functions. Remove rows with missing values in the 'price' column and convert 'date' to Date format"</p>
          </div>

          <h3 className="text-xl font-semibold mt-6 mb-3">Ask for Explanations</h3>
          <div className="bg-red-50 border-l-4 border-red-400 p-4 mb-2">
            <p className="font-semibold text-red-800">Less effective:</p>
            <p className="text-red-700 italic">"Use a linear mixed model"</p>
          </div>
          <div className="bg-green-50 border-l-4 border-green-400 p-4 mb-4">
            <p className="font-semibold text-green-800">More effective:</p>
            <p className="text-green-700 italic">"Can you explain why you chose to use a linear mixed model here instead of a standard linear regression?"</p>
          </div>

          {/* Getting Help */}
          <h2 id="getting-help" className="text-2xl font-semibold mt-10 mb-4">Getting Help</h2>
          <p className="mb-4">
            If you encounter issues or have questions:
          </p>
          <ul className="list-disc pl-6 mb-6 space-y-2">
            <li>Check the <a href="https://github.com/Omkar-Waingankar/rishi/issues" className="text-blue-600 hover:text-blue-800 underline">GitHub Issues</a> for known problems and solutions</li>
            <li>Open a new issue with detailed information about your problem</li>
            <li>Connect with the maintainer on <a href="https://www.linkedin.com/in/omkar-waingankar/" className="text-blue-600 hover:text-blue-800 underline">LinkedIn</a></li>
          </ul>

          {/* Privacy and Security */}
          <h2 id="privacy" className="text-2xl font-semibold mt-10 mb-4">Privacy and Security</h2>
          <p className="mb-4">
            Rishi is designed with privacy in mind:
          </p>
          <ul className="list-disc pl-6 mb-4 space-y-2">
            <li>All processing happens locally on your machine</li>
            <li>The messages you send are never used for training AI models</li>
            <li>Your code and data are never stored on external servers</li>
            <li>You control your own API keys and usage</li>
            <li>Rishi is fully open-source and auditable</li>
          </ul>
          <p className="mb-6">
            However, be aware that content you send to AI providers may be subject to their terms of service and privacy policies. Review the policies for <a href="https://www.anthropic.com/legal/privacy" className="text-blue-600 hover:text-blue-800 underline">Anthropic</a> and <a href="https://openai.com/policies/privacy-policy" className="text-blue-600 hover:text-blue-800 underline">OpenAI</a> if you have concerns.
          </p>
  

          <div className="mt-12 pt-8 border-t border-stone-200">
            <p className="text-center">
              <a href="/" className="text-blue-600 hover:text-blue-800 underline">
                ← Back to home
              </a>
            </p>
          </div>
        </article>
      </main>
    </div>
  );
}
