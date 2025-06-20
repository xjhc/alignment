# Walkthrough: Single VM Deployment on Azure

This document is the step-by-step, practical guide to deploying the `Alignment` application using our "Infrastructure as Script" philosophy. Follow these instructions to go from an empty Azure account to a running, live application.

### Prerequisites

*   An Azure account with an active subscription.
*   The [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli) installed and configured locally.
*   [jq](https://stedolan.github.io/jq/download/), a command-line JSON processor, installed locally.
*   Your project code pushed to a Git repository (e.g., GitHub).

---

## Part 1: Local Setup (One-Time)

You only need to do this once for your project.

1.  **Create `.env` Configuration:** This file holds all variables for your deployment. It is kept out of Git.
    ```bash
    # Copy the example file
    cp .env.example .env

    # Now, edit the new .env file with your specific values.
    ```

2.  **Ensure `.env` is in `.gitignore`:** Your root `.gitignore` file should contain this line:
    ```gitignore
    .env
    ```

3.  **Create the Scripts Directory:** Make a new directory `/scripts` at the project root and add the shell scripts provided at the end of this document.

4.  **Make Scripts Executable:** On Linux or macOS, grant execute permissions.
    ```bash
    chmod +x scripts/*.sh
    ```

---

## Part 2: The Deployment Workflow

Your entire workflow now revolves around three simple scripts run from your local machine.

*   **To Provision Infrastructure:**
    *   **Command:** `./scripts/provision.sh`
    *   **Action:** Creates the Azure Resource Group and the Virtual Machine. It saves the VM's IP address locally to `.azure_vm_ip` for other scripts to use. Run this once to set up your server.

*   **To Deploy or Update the Application:**
    *   **Command:** `./scripts/deploy.sh`
    *   **Action:** Connects to your VM, installs dependencies if needed, pulls the latest code from your Git repo, and launches/updates the application using Docker Compose. Run this anytime you push new code and want it live.

*   **To Destroy All Infrastructure:**
    *   **Command:** `./scripts/destroy.sh`
    *   **Action:** Permanently deletes the entire resource group and all resources within it. **This is irreversible and stops all Azure costs.**

---

## Part 3: Day-to-Day Operations

These helper scripts make it easy to manage and debug your running application.

*   **View Live Logs:** `./scripts/logs.sh`
*   **SSH into the VM:** `./scripts/ssh.sh`

---
