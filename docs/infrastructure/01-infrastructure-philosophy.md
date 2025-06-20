# Infrastructure: Philosophy & Strategy

This document outlines the high-level philosophy for deploying the `Alignment` application. Our goal for V1 is a strategy that prioritizes **simplicity, low cost, and reproducibility** over complexity and high availability.

## 1. The Chosen Approach: Infrastructure as Script

For our V1 deployment, we are adopting a middle-ground approach we call **"Infrastructure as Script."** This strategy is built on two core principles:

1.  **Scripts are the Source of Truth:** All commands needed to provision, deploy, and manage our infrastructure are captured in version-controlled shell scripts located in the `/scripts` directory. The repository itself becomes the definitive, executable guide to our deployment.
2.  **Configuration is Extracted:** All environment-specific variables (resource names, secrets, etc.) are stored in a local `.env` file. This file is explicitly excluded from version control via `.gitignore`, separating our non-sensitive scripts from our sensitive configuration.

### Why This Approach for V1?

This model provides the best balance of good practice and pragmatism for a small team or initial launch:
*   **Reproducibility:** Anyone on the team can provision an identical environment just by running `./scripts/provision.sh`.
*   **Simplicity:** It avoids the significant learning curve and complexity of full Infrastructure as Code (IaC) tools, which are overkill for a single VM setup.
*   **Low Cost:** It gives us direct control over a single, inexpensive VM, avoiding the higher costs associated with managed PaaS platforms.
*   **Audit Trail:** Every change to the deployment process is a commit to a script, tracked in our Git history.

## 2. Alternatives Considered

We considered several other common deployment models before choosing this path.

#### Alternative 1: Manual Deployment (via Azure Portal)
*   **Description:** Manually clicking through the Azure web portal to create a VM, then SSH'ing in to install dependencies and run the application.
*   **Pros:** Very low initial barrier to entry.
*   **Cons (Why we rejected it):** This "click-ops" approach is highly error-prone, impossible to reproduce consistently, and has no audit trail. It's fine for a quick, disposable test but unsuitable for a real project.

#### Alternative 2: Platform as a Service (PaaS)
*   **Description:** Using managed services like Azure App Service and Azure Cache for Redis. The cloud provider manages the OS, patching, and scaling.
*   **Pros:** Easiest to manage, excellent for scaling, no server maintenance.
*   **Cons (Why we deferred it):** Significantly more expensive than a single VM for a low-traffic application. It also offers less direct control over the environment. This is a great **future scaling path** but not the most cost-effective starting point.

#### Alternative 3: Full Infrastructure as Code (IaC) Tools
*   **Description:** Using a dedicated tool to define all cloud resources in a declarative format. This is the industry standard for managing complex cloud infrastructure. Examples include:
    *   **Terraform:** The market leader; uses its own domain-specific language (HCL).
    *   **Pulumi:** A modern alternative that uses general-purpose programming languages like Go, TypeScript, or Python.
    *   **Bicep:** A Microsoft-native language for deploying Azure resources that is simpler than ARM templates.
*   **Pros:** Extremely powerful, stateful, and provides a true, declarative picture of your infrastructure.
*   **Cons (Why we rejected it for V1):** These tools, while powerful, introduce a level of complexity that is not justified for our initial goal of deploying a single VM. They all have a learning curve and require a process for managing a "state file" (or an equivalent backend service) which tracks the infrastructure's current condition. The setup and cognitive overhead outweigh the benefits when the entire infrastructure consists of one server. They are the right choice for a future multi-server, high-availability architecture, but overkill for our simple V1 needs.

## 3. Conclusion

The **"Infrastructure as Script"** model is our goldilocks solution for V1. It provides the discipline and reproducibility of IaC without the overhead, at the lowest possible cost.

With this philosophy established, the next document, `02-single-vm-deployment.md`, provides the concrete, step-by-step guide to implement it.