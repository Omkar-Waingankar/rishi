export default function Home() {
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
          {/* Logo */}
          <div className="w-32 h-32 logo-gradient rounded-full mb-16 flex items-center justify-center">
            <span className="text-white font-bold text-7xl logo-text">R</span>
          </div>

          {/* Title */}
          <div className="flex items-center gap-4 mb-8">
            <h1 className="text-4xl font-bold">Rishi</h1>
            <a
              href="https://github.com/Omkar-Waingankar/rishi"
              className="inline-flex items-center gap-2 px-3 py-1.5 border border-stone-300 rounded-md text-md hover:bg-stone-50 transition-colors"
              target="_blank"
              rel="noopener noreferrer"
            >
              <svg width="16" height="16" viewBox="0 0 98 96" fill="currentColor">
                <path fillRule="evenodd" clipRule="evenodd" d="M48.854 0C21.839 0 0 22 0 49.217c0 21.756 13.993 40.172 33.405 46.69 2.427.49 3.316-1.059 3.316-2.362 0-1.141-.08-5.052-.08-9.127-13.59 2.934-16.42-5.867-16.42-5.867-2.184-5.704-5.42-7.17-5.42-7.17-4.448-3.015.324-3.015.324-3.015 4.934.326 7.523 5.052 7.523 5.052 4.367 7.496 11.404 5.378 14.235 4.074.404-3.178 1.699-5.378 3.074-6.6-10.839-1.141-22.243-5.378-22.243-24.283 0-5.378 1.94-9.778 5.014-13.2-.485-1.222-2.184-6.275.486-13.038 0 0 4.125-1.304 13.426 5.052a47 47 0 0 1 12.214-1.63c4.125 0 8.33.571 12.213 1.63 9.302-6.356 13.427-5.052 13.427-5.052 2.67 6.763.97 11.816.485 13.038 3.155 3.422 5.015 7.822 5.015 13.2 0 18.905-11.404 23.06-22.324 24.283 1.78 1.548 3.316 4.481 3.316 9.126 0 6.6-.08 11.897-.08 13.526 0 1.304.89 2.853 3.316 2.364 19.412-6.52 33.405-24.935 33.405-46.691C97.707 22 75.788 0 48.854 0" />
              </svg>
              <span>GitHub</span>
            </a>
          </div>

          {/* What is Rishi? */}
          <h2 className="text-base font-semibold mt-8 mb-4">What is Rishi?</h2>
          <p className="mb-4">
            Rishi is an open-source AI coding agent for R.
          </p>
          <ul className="list-disc pl-6 mb-4 space-y-2">
            <li>
              <strong>Context-aware:</strong> sees and controls your Console, Environment, and Plots
            </li>
            <li>
              <strong>Grounded in reality:</strong> responses based on your code and package documentation
            </li>
            <li>
              <strong>Privacy-first:</strong> runs locally, your messages/code/data stay on your machine
            </li>
            <li>
              <strong>Transparent:</strong> fully open-source and auditable, control your own API keys
            </li>
          </ul>
          <p className="mb-4">
            Works seamlessly as a RStudio Addin.
          </p>

          {/* Screenshot */}
          <div className="mb-8 flex justify-center">
            <img
              src="/rishiscreenshot.jpg"
              alt="Screenshot of Rishi in action"
              className="rounded-md shadow border border-stone-200 max-w-full"
              style={{ maxHeight: 400 }}
            />
          </div>

          {/* How do I get started? */}
          <h2 className="text-base font-semibold mt-8 mb-4">How do I get started?</h2>
          <p className="mb-4">
            Integrating Rishi into your R workflow is as simple as installing an R package and launching the Addin.
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
            You'll need to provide your own Anthropic API key to get started. If you don't have one already, you can get one from the <a href="https://console.anthropic.com/api-keys" className="text-blue-600 hover:text-blue-800 underline">Anthropic Console</a>.
          </p>

          {/* What models are supported? */}
          <h2 className="text-base font-semibold mt-8 mb-4">What models are supported?</h2>
          <p className="mb-4">
            Currently, Claude Sonnet 3.7 and Claude Sonnet 4. Expanding support to other frontier and open-source models is on the roadmap.
          </p>

          {/* How is this different from ChatGPT? */}
          <h2 className="text-base font-semibold mt-8 mb-4">How is this different from ChatGPT?</h2>
          <p className="mb-4">
            ChatGPT and similar tools lack awareness of your R project, data, and visualizations. To use them effectively, you must manually copy context back and forth from your IDE with every interaction.
          </p>
          <p className="mb-4">
            Rishi automatically delivers relevant context from your Files, Environment, Console, and Plots panes to the LLM with every message. It can also work autonomously with these panes to accomplish tasks without constant hand-holding.
          </p>

          {/* How are people using Rishi? */}
          <h2 className="text-base font-semibold mt-8 mb-4">How are people using Rishi?</h2>
          <p className="mb-4">
            Data scientists use Rishi to quickly ramp up and run exploratory data analyses, debug obscure console errors, polish visualizations, and learn unfamiliar packages. 
          </p>

          {/* How private is Rishi? */}
          <h2 className="text-base font-semibold mt-8 mb-4">How private is Rishi?</h2>
          <p className="mb-4">
            Very. By default, your messages, code, and data remain on your local machine. Rishi runs locally and doesn't store information in the cloud. You're also in control of your own API keys so there are no surprise costs or concerns about your work being used to train models.
          </p>

          {/* How can I ask questions or share feedback? */}
          <h2 className="text-base font-semibold mt-8 mb-4">How can I ask questions or share feedback?</h2>
          <p className="mb-4">
            You can ask questions or share feedback on{" "}
            <a
              href="https://github.com/Omkar-Waingankar/rishi/issues"
              className="text-blue-600 hover:text-blue-800 underline"
            >
              GitHub Issues
            </a>
            {" "}or connect with me (Omkar Waingankar) on{" "}
            <a
              href="https://www.linkedin.com/in/omkar-waingankar/"
              className="text-blue-600 hover:text-blue-800 underline"
            >
              LinkedIn
            </a>
            .
          </p>

        </article>
      </main>
    </div>
  );
}
