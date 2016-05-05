package main

import (
	"fmt"
	"gopkg.in/libgit2/git2go.v23"
	"errors"
	"time"
)

type GitControl struct {
	Conf  Config
	Cred  *git.Cred
	Repos []*git.Repository
}

func NewGitControl(conf Config, repos []*git.Repository) *GitControl {
	return &GitControl{Repos: repos, Conf: conf}
}

func (gc *GitControl) credentialsCallback(url string, username string, allowedTypes git.CredType) (git.ErrorCode, *git.Cred) {
	ret, cred := git.NewCredSshKey("git", gc.Conf.PublicKey, gc.Conf.PrivateKey, "")
	return git.ErrorCode(ret), &cred
}

func (gc *GitControl) certificateCheckCallback(cert *git.Certificate, valid bool, hostname string) git.ErrorCode {
	return 0
}

func (gc *GitControl) fetchOptions() *git.FetchOptions {
	fo := git.FetchOptions{
		RemoteCallbacks: git.RemoteCallbacks{
			CredentialsCallback:      gc.credentialsCallback,
			CertificateCheckCallback: gc.certificateCheckCallback,
		},
	}
	return &fo
}

func (gc *GitControl) pushOptions() *git.PushOptions {
	po := git.PushOptions{
		RemoteCallbacks: git.RemoteCallbacks{
			CredentialsCallback:      gc.credentialsCallback,
			CertificateCheckCallback: gc.certificateCheckCallback,
		},
	}
	return &po
}

func (gc *GitControl) signature() *git.Signature {
	return &git.Signature{
		Name: gc.Conf.Name,
		Email: gc.Conf.Email,
		When: time.Now(),
	}
}

func (gc *GitControl) PullAll() error {
	for _, repo := range gc.Repos {
		err := gc.PullOne(repo)
		if err != nil {
			fmt.Errorf("error pulling repo: %v", err)
			return err
		}
	}

	return nil
}

func (gc *GitControl) repoFromPath(path string) (*git.Repository, error) {
	for _, repo := range gc.Repos {
		fmt.Println(repo.Path()) // /.git/
		fmt.Println(path)  // /.bash_aliases this comparison is not equal
		if repo.Path() == path {
			return repo, nil
		}
	}

	return nil, errors.New("repo from path not found")
}

func (gc *GitControl) Update(path string) error {
	fmt.Printf("Updating... %s\n", path)
	repo, err := gc.repoFromPath(path)
	if err != nil {
		return err
	}
	fmt.Println("Adding...")
	tree, err := gc.Add(repo)
	if err != nil {
		return err
	}
	fmt.Println("Committing...")
	err = gc.Commit(tree, repo)
	if err != nil {
		return err
	}

	fmt.Println("Done...")
	return nil
}

func (gc *GitControl) Add(repo *git.Repository) (*git.Tree, error) {
	index, err := repo.Index()
	if err != nil {
		return nil, err
	}
	err = index.AddAll([]string{}, git.IndexAddDefault, nil)
	if err != nil {
		return nil, err
	}
	treeId, err := index.WriteTreeTo(repo)
	if err != nil {
		return nil, err
	}
	err = index.Write()
	if err != nil {
		return nil, err
	}
	tree, err := repo.LookupTree(treeId)
	if err != nil {
		return nil, err
	}
	return tree, nil
}

func (gc *GitControl) createCommitMessage() string{
	t := time.Now()
	return fmt.Sprint("%s %s %s", t.Month(), t.Day(), t.Year())
}

func (gc *GitControl) Commit(tree *git.Tree, repo *git.Repository) (error) {
	sig := gc.signature()
	message := gc.createCommitMessage()
	commitId, err := repo.CreateCommit("HEAD", sig, sig, message, tree)
	if err != nil {
		return err
	}
	fmt.Printf("commit: %s", commitId)
	remote, err := repo.Remotes.Lookup("origin")
	if err != nil {
		return err
	}
	err = remote.Push([]string{"refs/heads/master"}, gc.pushOptions())
	if err != nil {
		return err
	}

	return nil
}


// get a bit messy here
func (gc *GitControl) PullOne(repo *git.Repository) error {
	// lookup the origin
	remote, err := repo.Remotes.Lookup("origin")
	if err != nil {
		return err
	}

	// fetch origin
	if err := remote.Fetch([]string{}, gc.fetchOptions(), ""); err != nil {
		return err
	}

	// lookup master
	master, err := repo.References.Lookup("refs/remotes/origin/master")
	if err != nil {
		return err
	}

	remoteBranchID := master.Target()

	annotatedCommit, err := repo.AnnotatedCommitFromRef(master)
	if err != nil {
		return err
	}
	mergeHeads := make([]*git.AnnotatedCommit, 1)
	mergeHeads[0] = annotatedCommit
	analysis, _, err := repo.MergeAnalysis(mergeHeads)
	if err != nil {
		return err
	}

	head, err := repo.Head()
	if err != nil {
		return err
	}

	// everything is up to date.
	if analysis & git.MergeAnalysisUpToDate != 0 {
		return nil
	}else if analysis & git.MergeAnalysisFastForward != 0 {
		// Fast-forward changes
		// Get remote tree
		remoteTree, err := repo.LookupTree(remoteBranchID)
		if err != nil {
			return err
		}

		// Checkout
		if err := repo.CheckoutTree(remoteTree, nil); err != nil {
			return err
		}

		branchRef, err := repo.References.Lookup("refs/heads/master")
		if err != nil {
			return err
		}

		// Point branch to the object
		branchRef.SetTarget(remoteBranchID, "")
		if _, err := head.SetTarget(remoteBranchID, ""); err != nil {
			return err
		}
	}
	//todo: conflict not dealt with
	//else if analysis & git.MergeAnalysisNormal != 0 {
//    // Just merge changes
//    if err := repo.Merge([]*git.AnnotatedCommit{annotatedCommit}, nil, nil); err != nil {
//        return err
//    }
//    // Check for conflicts
//    index, err := repo.Index()
//    if err != nil {
//        return err
//    }
//
//    if index.HasConflicts() {
//        return errors.New("Conflicts encountered. Please resolve them.")
//    }
//
//    // Make the merge commit
//    sig, err := repo.DefaultSignature()
//    if err != nil {
//        return err
//    }
//
//    // Get Write Tree
//    treeId, err := index.WriteTree()
//    if err != nil {
//        return err
//    }
//
//    tree, err := repo.LookupTree(treeId)
//    if err != nil {
//        return err
//    }
//
//    localCommit, err := repo.LookupCommit(head.Target())
//    if err != nil {
//        return err
//    }
//
//    remoteCommit, err := repo.LookupCommit(remoteBranchID)
//    if err != nil {
//        return err
//    }
//
//    repo.CreateCommit("HEAD", sig, sig, "", tree, localCommit, remoteCommit)
//
//    // Clean up
//    repo.StateCleanup()
//}



	return nil
}