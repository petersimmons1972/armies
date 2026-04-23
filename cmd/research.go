package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/petersimmons1972/armies/internal/config"
	"github.com/spf13/cobra"
)

// NewResearchCommand returns the research cobra.Command.
// It generates a research prompt document for creating a new agent profile.
func NewResearchCommand() *cobra.Command {
	var profilesDir string
	var mode string

	cmd := &cobra.Command{
		Use:   "research <role>",
		Short: "Generate a research prompt document for creating a new agent profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			role := args[0]

			// Validate mode
			switch mode {
			case "prompt":
				// default — write file
			case "api":
				fmt.Fprintln(cmd.ErrOrStderr(), "API mode not yet implemented. Default prompt mode used.")
			default:
				return fmt.Errorf("invalid --mode %q: must be \"prompt\" or \"api\"", mode)
			}

			// Resolve profiles dir — fall back to config if flag not set
			pdir := profilesDir
			if pdir == "" {
				cfg, err := config.Load()
				if err != nil {
					return err
				}
				pdir, err = cfg.ProfilesDirLexicallyValidated()
				if err != nil {
					return err
				}
			}

			// Create drafts subdirectory
			draftsDir := filepath.Join(pdir, "drafts")
			if err := os.MkdirAll(draftsDir, 0o700); err != nil {
				return fmt.Errorf("cannot create drafts directory %s: %w", draftsDir, err)
			}

			today := time.Now().Format("2006-01-02")
			filename := "draft-" + role + "-" + today + ".md"
			draftPath := filepath.Join(draftsDir, filename)

			content := buildResearchPrompt(role, today)
			if err := os.WriteFile(draftPath, []byte(content), 0o600); err != nil {
				return fmt.Errorf("cannot write draft file %s: %w", draftPath, err)
			}

			// Print using profiles/ prefix for readability
			relPath := filepath.Join("profiles", "drafts", filename)
			fmt.Fprintf(cmd.OutOrStdout(), "Draft prompt saved to %s\n", relPath)
			fmt.Fprintln(cmd.OutOrStdout(), "Feed this file to a Claude Code agent using the Agent tool to generate a complete profile.")

			return nil
		},
	}

	cmd.Flags().StringVar(&profilesDir, "profiles-dir", "", "Profiles directory")
	cmd.Flags().StringVar(&mode, "mode", "prompt", "Generation mode: prompt or api")

	return cmd
}

// buildResearchPrompt generates the full research prompt document for the given role.
func buildResearchPrompt(role, today string) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "# Research Prompt: %s Agent Profile\n", role)
	fmt.Fprintf(&sb, "Generated: %s\n", today)
	sb.WriteString("\n")
	sb.WriteString("## What You Are Building\n")
	sb.WriteString("\n")
	sb.WriteString("An activation profile — not a biography, not a summary, not a list of achievements.\n")
	sb.WriteString("The model you are running on already knows every well-documented historical figure\n")
	sb.WriteString("deeply. Your job is not to teach it who the person is. Your job is to find the\n")
	sb.WriteString("specific behavioral pointers that unlock the right slice of that knowledge and focus\n")
	fmt.Fprintf(&sb, "it on the `%s` role.\n", role)
	sb.WriteString("\n")
	sb.WriteString("A profile that describes achievements produces a generic agent.\n")
	sb.WriteString("A profile that describes how someone moved, decided, failed, and recovered\n")
	sb.WriteString("produces a useful one.\n")
	sb.WriteString("\n")
	sb.WriteString("## Step 1 — Select the Figure\n")
	sb.WriteString("\n")
	fmt.Fprintf(&sb, "Research 3–5 real historical figures who naturally fit the `%s` role class.\n", role)
	sb.WriteString("\n")
	sb.WriteString("Selection criteria — in order of importance:\n")
	sb.WriteString("1. **Documentation depth**: Has history argued about this person for decades?\n")
	sb.WriteString("   Are there multiple biographies that disagree with each other? Primary sources\n")
	sb.WriteString("   (letters, memoirs, contemporaries' accounts)? The richer the record, the\n")
	sb.WriteString("   stronger the activation. Avoid figures known primarily through one source.\n")
	sb.WriteString("2. **Behavioral specificity**: Do we know HOW they worked, not just WHAT they\n")
	sb.WriteString("   achieved? How they ran a meeting. How they delivered bad news. How they made\n")
	sb.WriteString("   decisions under uncertainty. How they failed specifically.\n")
	fmt.Fprintf(&sb, "3. **Role fit**: Do their documented working patterns naturally map to `%s`\n", role)
	sb.WriteString("   behaviors — not just their job title or reputation?\n")
	sb.WriteString("\n")
	sb.WriteString("For each candidate note:\n")
	sb.WriteString("- Why the documentation record is deep enough to activate well\n")
	fmt.Fprintf(&sb, "- One specific behavioral detail (not an achievement) that maps to `%s`\n", role)
	sb.WriteString("- Their specific, documented failure mode — not a generic weakness\n")
	sb.WriteString("\n")
	sb.WriteString("## Step 2 — Select the Best Candidate\n")
	sb.WriteString("\n")
	sb.WriteString("Choose the single best candidate. Prioritize documentation depth over fame.\n")
	sb.WriteString("A less famous figure with a rich behavioral record outperforms a famous figure\n")
	sb.WriteString("with a thin one. Explain why this figure's documented behavioral patterns\n")
	fmt.Fprintf(&sb, "make them the strongest activation key for the `%s` role.\n", role)
	sb.WriteString("\n")
	sb.WriteString("## Step 3 — Research the Second Layer\n")
	sb.WriteString("\n")
	sb.WriteString("Before writing the profile, do the deep research. Do not write from memory.\n")
	sb.WriteString("Search for:\n")
	sb.WriteString("\n")
	sb.WriteString("- How they actually ran a room — specific documented behaviors, not general reputation\n")
	sb.WriteString("- The decisions they made under pressure that reveal character\n")
	sb.WriteString("- Contemporaries' accounts — what did people who worked with them say?\n")
	sb.WriteString("- The specific failure that cost them something real and documented\n")
	sb.WriteString("- The thing they did that surprised people who expected something different\n")
	sb.WriteString("- Anything that contradicts the surface reputation\n")
	sb.WriteString("\n")
	sb.WriteString("The surface reputation is what everyone already knows. The model already has it.\n")
	sb.WriteString("You are looking for the second layer — the behavioral detail that is documented\n")
	sb.WriteString("but not famous. That is what makes the profile work.\n")
	sb.WriteString("\n")
	sb.WriteString("## Step 4 — Write the Profile\n")
	sb.WriteString("\n")
	sb.WriteString("Use the following format:\n")
	sb.WriteString("\n")
	sb.WriteString("```\n")
	sb.WriteString("---\n")
	sb.WriteString("name: [kebab-case]\n")
	sb.WriteString("display_name: \"[Full Title and Name]\"\n")
	sb.WriteString("description: >\n")
	sb.WriteString("  [3-4 sentences — behavioral description for THIS role specifically.\n")
	sb.WriteString("   Not achievements. Not reputation. How they operate and when to use them.]\n")
	sb.WriteString("roles:\n")
	fmt.Fprintf(&sb, "  primary: %s\n", role)
	sb.WriteString("  secondary: [if applicable]\n")
	sb.WriteString("xp: 0\n")
	sb.WriteString("rank: \"[Historical rank or title]\"\n")
	sb.WriteString("model: [opus for coordinator/planner/researcher; sonnet for implementer/troubleshooter]\n")
	sb.WriteString("[disallowedTools: — only if coordinator role]\n")
	sb.WriteString("[  - Write]\n")
	sb.WriteString("[  - Edit]\n")
	sb.WriteString("[  - Bash]\n")
	sb.WriteString("---\n")
	sb.WriteString("\n")
	sb.WriteString("## Base Persona\n")
	sb.WriteString("\n")
	sb.WriteString("[300-400 words of behavioral prose. Write in second person (\"You are...\").\n")
	sb.WriteString(" Include:\n")
	sb.WriteString(" - Formation: what made them who they are. Specific, not generic.\n")
	sb.WriteString(" - How they actually work: the specific behaviors that distinguish them\n")
	sb.WriteString("   from a type. Not \"decisive\" but what decisive looked like for this person.\n")
	sb.WriteString(" - A specific relationship or training experience that shaped their method.\n")
	sb.WriteString(" - **Named failure mode**: one specific, documented failure with real\n")
	sb.WriteString("   consequences — not a character flaw, a real thing that happened.\n")
	sb.WriteString("   This is load-bearing. It creates accountability and makes the agent\n")
	sb.WriteString("   feel real rather than oracular.\n")
	sb.WriteString(" - One behavioral detail that contradicts or complicates the surface reputation.]\n")
	sb.WriteString("\n")
	fmt.Fprintf(&sb, "## Role: %s\n", role)
	sb.WriteString("\n")
	sb.WriteString("[150-200 words of operational instructions for this specific role.\n")
	sb.WriteString(" Pre-mission checklist. How they work. What they deliver. What \"done\" looks like.\n")
	sb.WriteString(" These are not generic role instructions — they are how THIS person approaches\n")
	sb.WriteString(" THIS role based on their documented working patterns.]\n")
	sb.WriteString("```\n")
	sb.WriteString("\n")
	sb.WriteString("## Step 5 — Write the Behavioral Fingerprints\n")
	sb.WriteString("\n")
	sb.WriteString("After the profile body, add a `test_scenarios` block to the frontmatter.\n")
	sb.WriteString("This is how you verify the profile is activating the right person, not\n")
	sb.WriteString("producing a generic agent with the correct name.\n")
	sb.WriteString("\n")
	sb.WriteString("Write exactly 3 scenarios using these archetypes — every profile gets all three:\n")
	sb.WriteString("\n")
	sb.WriteString("1. **ambiguous-order**: A task with a missing constraint. Watch how they seek clarity.\n")
	sb.WriteString("2. **pressure-test**: A deadline compressed mid-campaign. Watch how they push back.\n")
	sb.WriteString("3. **scope-creep-trap**: An out-of-scope request added mid-campaign. Watch the negotiation.\n")
	sb.WriteString("\n")
	sb.WriteString("For each scenario, write 2 fingerprint criteria. Each criterion must:\n")
	sb.WriteString("- Describe a specific behavior you would expect from THIS person that a generic\n")
	fmt.Fprintf(&sb, "  `%s` agent would never produce\n", role)
	sb.WriteString("- Include a `why` field explaining what the generic version looks like and why\n")
	sb.WriteString("  this person's documented behavior differs — cite the specific research that\n")
	sb.WriteString("  supports it (a relationship, an incident, a documented habit)\n")
	sb.WriteString("\n")
	sb.WriteString("Format:\n")
	sb.WriteString("\n")
	sb.WriteString("```yaml\n")
	sb.WriteString("test_scenarios:\n")
	sb.WriteString("  - id: ambiguous-order\n")
	sb.WriteString("    situation: >\n")
	sb.WriteString("      [2-3 sentences describing the situation. Make it realistic and role-appropriate.]\n")
	sb.WriteString("    prompt: \"[The single question or instruction the agent must respond to.]\"\n")
	sb.WriteString("    fingerprints:\n")
	sb.WriteString("      - criterion: [Specific behavior expected — one sentence, observable]\n")
	sb.WriteString("        why: >\n")
	sb.WriteString("          [What the generic version looks like. Why this person's documented\n")
	sb.WriteString("           history produces a different response. Cite the specific source —\n")
	sb.WriteString("           an incident, a relationship, a documented working habit.]\n")
	sb.WriteString("      - criterion: [Second behavior]\n")
	sb.WriteString("        why: >\n")
	sb.WriteString("          [Same structure.]\n")
	sb.WriteString("  - id: pressure-test\n")
	sb.WriteString("    [same structure]\n")
	sb.WriteString("  - id: scope-creep-trap\n")
	sb.WriteString("    [same structure]\n")
	sb.WriteString("```\n")
	sb.WriteString("\n")
	sb.WriteString("The `why` field is load-bearing. Without it, the rubric is just a checklist.\n")
	sb.WriteString("With it, the person scoring the test knows exactly what they are listening for\n")
	sb.WriteString("and why a generic response fails.\n")
	sb.WriteString("\n")
	sb.WriteString("Bad fingerprint (too generic):\n")
	sb.WriteString("  criterion: \"Asks clarifying questions before proceeding\"\n")
	sb.WriteString("  why: \"Good coordinators ask questions.\"\n")
	sb.WriteString("\n")
	sb.WriteString("Good fingerprint (specific to person):\n")
	sb.WriteString("  criterion: \"Names the missing constraint before issuing any assignments\"\n")
	sb.WriteString("  why: >\n")
	sb.WriteString("    \"A generic coordinator assumes or asks vaguely. Eisenhower's documented habit —\n")
	sb.WriteString("     from his Abilene poker education through every command — was to write down\n")
	sb.WriteString("     what he did not know before committing. He would not brief specialists on an\n")
	sb.WriteString("     ambiguous order. If the response assigns work without naming the gap, this fails.\"\n")
	sb.WriteString("\n")
	sb.WriteString("## Step 6 — Verify Before Saving\n")
	sb.WriteString("\n")
	sb.WriteString("Read the Base Persona back. Ask:\n")
	sb.WriteString("- Does this feel like a specific person or a type?\n")
	sb.WriteString("- Could this description apply to three other people in the same role? If yes, it is too generic.\n")
	sb.WriteString("- Is the failure mode a real documented event with real consequences? Or a character note?\n")
	sb.WriteString("- Does the description tell you HOW they moved, or just WHAT they achieved?\n")
	sb.WriteString("\n")
	sb.WriteString("Read the fingerprints. Ask:\n")
	fmt.Fprintf(&sb, "- Would a generic %s agent produce this behavior? If yes, the fingerprint is too weak.\n", role)
	sb.WriteString("- Does the `why` field cite a specific documented behavior, or does it just explain the criterion?\n")
	sb.WriteString("- Could you score a response against this criterion, or is it too vague to judge?\n")
	sb.WriteString("\n")
	sb.WriteString("If the answers are wrong, research more before saving.\n")
	sb.WriteString("\n")
	sb.WriteString("## Step 7 — Save and Commit\n")
	sb.WriteString("\n")
	sb.WriteString("Save the completed profile to:\n")
	sb.WriteString("\n")
	sb.WriteString("    ~/.armies/profiles/<name>.md\n")
	sb.WriteString("\n")
	sb.WriteString("where `<name>` is the agent's lowercase hyphenated identifier.\n")
	sb.WriteString("\n")
	sb.WriteString("Verify it works: `armies test <name>` should print the full test document without errors.\n")
	sb.WriteString("\n")
	fmt.Fprintf(&sb, "Commit: `git -C ~/.armies commit -am \"profile(<name>): %s role — [figure name]\"`\n", role)
	sb.WriteString("\n")
	sb.WriteString("## Hard Constraints\n")
	sb.WriteString("\n")
	sb.WriteString("- Real historical figures only. No fictional characters — they lack the\n")
	sb.WriteString("  multi-source documentation depth that produces strong activation.\n")
	sb.WriteString("- Do NOT write from the Wikipedia lede. That is the surface. Go deeper.\n")
	sb.WriteString("- Do NOT re-use figures already in ~/.armies/profiles/.\n")
	sb.WriteString("- Do NOT start at 0 words on the failure mode. Every profile needs one.\n")
	sb.WriteString("- Do NOT skip test_scenarios. Every profile must pass `armies test` on creation.\n")
	sb.WriteString("- The profile must pass `armies roster` without errors after saving.\n")

	return sb.String()
}

func init() {
	rootCmd.AddCommand(NewResearchCommand())
}
