package main

import (
    "fmt"
    "github.hpe.com/christophe-larsonneur/goforjj/trace"
    "os"
    "os/exec"
    "path"
    "strings"
    "io/ioutil"
    "path/filepath"
    "github.hpe.com/christophe-larsonneur/goforjj"
)

// Ensure local repo exist with at least 1 commit.
// If non existent, or no commit, it will create it all.
func (a *Forj) ensure_local_repo(repo_name string) error {
    repo := path.Clean(path.Join(a.Workspace_path, a.Workspace, repo_name))

    gotrace.Trace("Checking '%s' repository...", repo)
    dir, err := os.Stat(repo)
    if os.IsNotExist(err) {
        if git("init", repo) > 0 {
            return fmt.Errorf("Unable to initialize %s\n", repo)
        }
        gotrace.Trace("Created '%s' repository...", repo)
    }

    dir, err = os.Stat(path.Join(repo, ".git"))
    if os.IsNotExist(err) {
        gotrace.Trace("Existing directory '%s' will became a git repo", repo)
        if git("init", repo) > 0 {
            return fmt.Errorf("Unable to initialize %s\n", repo)
        }
        gotrace.Trace("Initialized '%s' directory as git repository...", repo)
    }

    if os.IsExist(err) && !dir.IsDir() {
        return fmt.Errorf("'%s' is not a valid GIT repository. Please fix it first. '%s' is not a directory.\n", repo, path.Join(repo, ".git"))
    }

    // if _, err := git_get("config", "--get", "remote.origin.url"); err != nil {
    if err := os.Chdir(repo) ; err != nil {
        fmt.Printf("Unable to move to '%s' : %s\n", err)
        os.Exit(1)
    }
    if found, err := git_get("config", "--local", "-l"); err != nil {
        return fmt.Errorf("'%s' is not a valid GIT repository. Please fix it first. %s\n", repo, err)
    } else {
        gotrace.Trace("Valid local git config found: \n%s", found)
    }

    if _, err := git_get("log", "-1"); err != nil {
        git_1st_commit(repo)
        gotrace.Trace("Initial commit created.")
    } else {
        gotrace.Trace("nothing done on existing '%s' git repository...", repo)
    }
    return nil
}

// Do commit from data returned by a plugin.
func (a *Forj) DoCommit(d *goforjj.PluginData) error {
    gotrace.Trace("Committing %d files as '%s'", len(d.Files), d.CommitMessage)
    if len(d.Files) == 0 {
        return fmt.Errorf("Nothing to commit")
    }

    if d.CommitMessage == "" {
        return fmt.Errorf("Unable to commit without a commit message.")
    }

    for _, file := range d.Files {
        if i := git("add", path.Join("apps", a.CurrentPluginDriver.driver_type, file)); i >0 {
            return fmt.Errorf("Issue while adding code to git. RC=%d", i)
        }
    }
    git("commit", "-m", d.CommitMessage)
    return nil
}

// Call git command with arguments. All print out displayed. It returns git Return code.
func git(opts ...string) int {
    return run_cmd("git", opts...)
}

// Call a git command and get the output as string output.
func git_get(opts ...string) (string, error) {
    gotrace.Trace("RUNNING: git %s", strings.Join(opts, " "))
    out, err := exec.Command("git", opts...).Output()
    return string(out), err
}

// Create initial commit
func git_1st_commit(repo string) {
    readme_path := path.Join(repo, "README.md")

    // check if an existing README exist to keep
    _, err := os.Stat(readme_path)
    if ! os.IsNotExist(err) {
        gotrace.Trace("Renaming '%s' to '%s'", readme_path, readme_path + ".forjj_tmp")
        if err := os.Rename(readme_path, readme_path + ".forjj_tmp") ; err!= nil {
            fmt.Printf("Unable to rename '%s' to '%s'. %s\n", readme_path, readme_path + ".forjj_tmp", err)
            os.Exit(1)
        }
    }

    // Generate README.md
    // TODO: Support for a template data instead.
    gotrace.Trace("Generate %s", readme_path)
    data := []byte(fmt.Sprintf("FYI: This project has been generated by forjj\n\n%s Infra Repository\n", filepath.Base(repo)))
    if err := ioutil.WriteFile(readme_path, data, 0644) ; err!= nil {
        fmt.Printf("Unable to create '%s'. %s\n", readme_path, err)
        os.Exit(1)
    }

    git("add", "README.md")
    git("commit", "-m", "Initial commit")

    // check if an original README.md was there to restore his content.
    _, err = os.Stat(readme_path + ".forjj_tmp")
    if ! os.IsNotExist(err) {
        gotrace.Trace("Renaming '%s' to '%s'", readme_path + ".forjj_tmp", readme_path)
        if err := os.Rename(readme_path + ".forjj_tmp", readme_path) ; err!= nil {
            fmt.Printf("Unable to rename '%s' to '%s'. %s\n", readme_path + ".forjj_tmp", readme_path, err)
            os.Exit(1)
        }
    }
}

