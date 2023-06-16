package gitclient

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

func PrepareDestinationDirectory(dir string) (*git.Repository, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.Mkdir(dir, 0755)
		if err != nil {
			return nil, err
		}
	}

	if _, err := os.Stat(dir + "/.git"); os.IsNotExist(err) {
		repo, err := git.PlainInit(dir, false)
		if err != nil {
			return nil, err
		}
		return repo, err
	}
	repo, err := git.PlainOpen(dir)
	if err != nil {
		return nil, err
	}
	// Check if the branch "head" already exists
	headRefName := plumbing.NewBranchReferenceName("head")
	_, err = repo.Reference(headRefName, true)
	if err == nil {
		fmt.Println("Branch 'head' already exists.")
		return repo, nil
	}

	// Resolve the HEAD reference
	headRef, err := repo.Head()
	if err != nil {
		fmt.Println("Failed to get HEAD reference:", err)
		return nil, err
	}

	// Create a new branch reference pointing to the HEAD reference
	branchRef := plumbing.NewHashReference(headRefName, headRef.Hash())

	// Set the branch reference in the repository
	err = repo.Storer.SetReference(branchRef)
	if err != nil {
		fmt.Println("Failed to create new branch 'head':", err)
		return nil, err
	}

	// Set the HEAD reference in the configuration
	cfg, err := repo.Config()
	if err != nil {
		fmt.Println("Failed to get repository configuration:", err)
		return nil, err
	}
	cfg.Core.IsBare = false
	err = repo.Storer.SetConfig(cfg)
	if err != nil {
		fmt.Println("Failed to set HEAD reference in configuration:", err)
		return nil, err
	}

	fmt.Println("New HEAD reference 'head' created successfully.")
	return repo, nil
}

func CommitAllChanges(repo *git.Repository) error {
	worktree, err := repo.Worktree()
	if err != nil {
		fmt.Println("error while getting worktree")
		return err
	}
	// err = createMasterBranchIfNotExists(repo)
	// if err != nil {
	// 	fmt.Println("error while creating master")
	// 	return err
	// }

	// Add all changes
	err = worktree.AddGlob("*")
	if err != nil {
		fmt.Println("errror while adding changes")
		return err
	}

	// Commit changes
	_, err = worktree.Commit("change applied", &git.CommitOptions{
		All: true,
		Author: &object.Signature{
			Name:  "Byteblaze",
			Email: "byteblaze@akamai.com",
			When:  time.Now(),
		},
	})

	if err != nil {
		fmt.Println("error while commiting")
		return err
	}
	return nil
}

func RollbackChanges(repo *git.Repository) error {
	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	err = w.Checkout(&git.CheckoutOptions{
		Hash:  plumbing.NewHash("HEAD"),
		Force: true,
	})
	if err != nil {
		return err
	}
	return nil
}

func createMasterBranchIfNotExists(repo *git.Repository) error {
	// Check if the branch already exists
	branchRef := plumbing.NewBranchReferenceName("master")
	_, err := repo.Reference(branchRef, true)
	if err == nil {
		fmt.Println("Branch already exists:", branchRef)
		return err
	}
	// Get repository's storage
	storer := repo.Storer

	// Get HEAD reference
	headRef, err := storer.Reference(plumbing.HEAD)
	if err != nil {
		fmt.Println("error while getting head")
		return err
	}

	// Create a new branch based on the HEAD reference
	newBranch := plumbing.NewHashReference(branchRef, headRef.Hash())

	// Update the reference in the repository to point to the new branch
	err = repo.Storer.SetReference(newBranch)
	if err != nil {
		fmt.Println("Failed to create new branch:", err)
		return err
	}
	fmt.Println("Branch created")
	return nil
}
