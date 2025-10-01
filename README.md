# Rishi

**Rishi is an open-source AI coding agent for R:**

- **Context-aware:** sees and controls your Console, Environment, and Plots
- **Grounded in reality:** responses based on your code and package documentation
- **Privacy-first:** runs locally, your messages/code/data stay on your machine
- **Transparent:** fully open-source and auditable, control your own API keys

Works seamlessly as a RStudio Addin.

<table>
  <tr>
    <td width="30%" valign="top"><b>What is Rishi?</b></td>
    <td width="70%" valign="top">Rishi is an open-source AI coding agent for R that integrates seamlessly into RStudio. Unlike ChatGPT and similar tools, Rishi sees your R project, data, and visualizations — no more copy-pasting context back and forth.</td>
  </tr>
  <tr>
    <td width="30%" valign="top"><b>How is Rishi different from ChatGPT?</b></td>
    <td width="70%" valign="top">
      ChatGPT and similar tools lack awareness of your R project, data, and visualizations. To use them effectively, you must manually copy context back and forth from your IDE with every interaction.<br><br>
      Rishi automatically delivers relevant context from your Files, Environment, Console, and Plots panes to the LLM with every message. It can also work autonomously with these panes to accomplish tasks without constant hand-holding.
    </td>
  </tr>
  <tr>
    <td width="30%" valign="top"><b>What models are supported?</b></td>
    <td width="70%" valign="top">Currently, Claude Sonnet 3.7 and Claude Sonnet 4. Expanding support to other frontier and open-source models is on the roadmap.</td>
  </tr>
  <tr>
    <td width="30%" valign="top"><b>How private is Rishi?</b></td>
    <td width="70%" valign="top">Very. By default, your messages, code, and data remain on your local machine. Rishi runs locally and doesn't store information in the cloud. You're also in control of your own API keys so there are no surprise costs or concerns about your work being used to train models.</td>
  </tr>
  <tr>
    <td width="30%" valign="top"><b>How are people using Rishi?</b></td>
    <td width="70%" valign="top">Data scientists use Rishi to quickly ramp up and run exploratory data analyses, debug obscure console errors, polish visualizations, and learn unfamiliar packages.</td>
  </tr>
  <tr>
    <td width="30%" valign="top"><b>How do I get started?</b></td>
    <td width="70%" valign="top">Integrating Rishi into your R workflow is as simple as installing an R package and launching the Addin. See the <b><a href="#get-started">Quick Start</a></b> below.</td>
  </tr>
</table>

---

## Get Started

**Install Rishi:**

```r
# Install remotes if you don't have it
install.packages("remotes")

# Install Rishi
remotes::install_github("Omkar-Waingankar/rishi", subdir = "addin")
```

**Launch Rishi:**

```r
rishi::rishiAddin()
```

Or access it from the RStudio Addins menu.

---

## Questions & Feedback

You can ask questions or share feedback on [GitHub Issues](https://github.com/Omkar-Waingankar/rishi/issues) or connect with me (Omkar Waingankar) on [LinkedIn](https://www.linkedin.com/in/omkar-waingankar/).

## Contributing

Want to contribute to Rishi? Check out our [Contributing Guide](CONTRIBUTING.md) for development setup, build instructions, and release procedures.
