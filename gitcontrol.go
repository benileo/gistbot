package main

import (
	"errors"
	"fmt"
	"gopkg.in/libgit2/git2go.v23"
	"log"
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

func (gc *GitControl) PullAll() {
	ch := make(chan error)

	for _, repo := range gc.Repos {
		go gc.PullOne(repo, ch)
	}

	for i := 0; i < len(gc.Repos); i++ {
		err := <-ch
		if err != nil {
			fmt.Errorf("error pulling repo: %v", err)
			log.Println(err)
		}
	}
}

func (gc *GitControl) Origin(repo *git.Repository) (*git.Remote, error) {
	remote, err := repo.Remotes.Lookup("origin")
	if err != nil {
		fmt.Errorf("error looking up origin: %v", err)
		return nil, err
	}

	return remote, nil
}

func (gc *GitControl) Master(repo *git.Repository) (*git.Reference, error) {
	master, err := repo.References.Lookup("refs/remotes/origin/master")
	if err != nil {
		fmt.Errorf("error looking up master branch: %v", err)
		return nil, err
	}

	return master, nil
}

func (gc *GitControl) FastForward(repo *git.Repository, oid *git.Oid) error{
	head, err := repo.Head()
	if err != nil {
		return fmt.Errorf("error getting HEAD: %v", err)
	}

	// Lookup the git tree object for the given ref
	remoteTree, err := repo.LookupTree(oid)
	if err != nil {
		return fmt.Errorf("error lookup tree HEAD: %v", err)
	}

	// Honestly, don't fully understand this one
	if err := repo.CheckoutTree(remoteTree, nil); err != nil {
		return err
	}

	// lookup local master branch
	branchRef, err := repo.References.Lookup("refs/heads/master")
	if err != nil {
		return fmt.Errorf("unable to lookup refs/heads/master: %v", err)
	}

	// Point branch to the object (don't fully understand this one too
	branchRef.SetTarget(oid, "")
	if _, err := head.SetTarget(oid, ""); err != nil {
		return err
	}

	return nil
}

func (gc *GitControl) Fetch(repo *git.Repository) error {
	remote, err := gc.Origin(repo)
	if err != nil {
		return err
	}

	if err = remote.Fetch([]string{}, gc.fetchOptions(), ""); err != nil {
		return err
	}

	return nil
}

func (gc *GitControl) Merge(repo *git.Repository) error {
	masterRemote, err := gc.Master(repo)
	if err != nil {
		return err
	}

	target := masterRemote.Target()
	annotatedCommit, _ := repo.AnnotatedCommitFromRef(masterRemote)
	mergeHeads := []*git.AnnotatedCommit{annotatedCommit}
	analysis, _, _ := repo.MergeAnalysis(mergeHeads)

	if analysis & git.MergeAnalysisUpToDate != 0 {
		log.Println("Everything up to date")

		return nil
	} else if analysis & git.MergeAnalysisFastForward != 0 {
		// Fast forward

		if err = gc.FastForward(repo, target); err != nil {
			return err
		}
	}

	return nil
}

func (gc *GitControl) PullOne(repo *git.Repository, ch chan error) {
	if err := gc.Fetch(repo); err != nil {
		ch <- err
		return
	}

	if err := gc.Merge(repo); err != nil {
		ch <- err
		return
	}

	ch <- nil
}


//func (gc *GitControl) signature() *git.Signature {
//	return &git.Signature{
//		Name:  gc.Conf.Name,
//		Email: gc.Conf.Email,
//		When:  time.Now(),
//	}
//}
//
//func (gc *GitControl) repoFromPath(path string) (*git.Repository, error) {
//	for _, repo := range gc.Repos {
//		fmt.Println(repo.Path()) // /.git/
//		fmt.Println(path)        // /.bash_aliases this comparison is not equal
//		if repo.Path() == path {
//			return repo, nil
//		}
//	}
//
//	return nil, errors.New("repo from path not found")
//}
//
//func (gc *GitControl) Update(path string) error {
//	fmt.Printf("Updating... %s\n", path)
//	repo, err := gc.repoFromPath(path)
//	if err != nil {
//		return err
//	}
//	fmt.Println("Adding...")
//	tree, err := gc.Add(repo)
//	if err != nil {
//		return err
//	}
//	fmt.Println("Committing...")
//	err = gc.Commit(tree, repo)
//	if err != nil {
//		return err
//	}
//
//	fmt.Println("Done...")
//	return nil
//}
//
//func (gc *GitControl) Add(repo *git.Repository) (*git.Tree, error) {
//	index, err := repo.Index()
//	if err != nil {
//		return nil, err
//	}
//	err = index.AddAll([]string{}, git.IndexAddDefault, nil)
//	if err != nil {
//		return nil, err
//	}
//	treeId, err := index.WriteTreeTo(repo)
//	if err != nil {
//		return nil, err
//	}
//	err = index.Write()
//	if err != nil {
//		return nil, err
//	}
//	tree, err := repo.LookupTree(treeId)
//	if err != nil {
//		return nil, err
//	}
//	return tree, nil
//}
//
//func (gc *GitControl) createCommitMessage() string {
//	t := time.Now()
//	return fmt.Sprint("%s %s %s", t.Month(), t.Day(), t.Year())
//}
//
//func (gc *GitControl) Commit(tree *git.Tree, repo *git.Repository) error {
//	sig := gc.signature()
//	message := gc.createCommitMessage()
//	commitId, err := repo.CreateCommit("HEAD", sig, sig, message, tree)
//	if err != nil {
//		return err
//	}
//	fmt.Printf("commit: %s", commitId)
//	remote, err := repo.Remotes.Lookup("origin")
//	if err != nil {
//		return err
//	}
//	err = remote.Push([]string{"refs/heads/master"}, gc.pushOptions())
//	if err != nil {
//		return err
//	}
//
//	return nil
//}

////todo: conflict not dealt with
////else if analysis & git.MergeAnalysisNormal != 0 {
////    // Just merge changes
////    if err := repo.Merge([]*git.AnnotatedCommit{annotatedCommit}, nil, nil); err != nil {
////        return err
////    }
////    // Check for conflicts
////    index, err := repo.Index()
////    if err != nil {
////        return err
////    }
////
////    if index.HasConflicts() {
////        return errors.New("Conflicts encountered. Please resolve them.")
////    }
////
////    // Make the merge commit
////    sig, err := repo.DefaultSignature()
////    if err != nil {
////        return err
////    }
////
////    // Get Write Tree
////    treeId, err := index.WriteTree()
////    if err != nil {
////        return err
////    }
////
////    tree, err := repo.LookupTree(treeId)
////    if err != nil {
////        return err
////    }
////
////    localCommit, err := repo.LookupCommit(head.Target())
////    if err != nil {
////        return err
////    }
////
////    remoteCommit, err := repo.LookupCommit(remoteBranchID)
////    if err != nil {
////        return err
////    }
////
////    repo.CreateCommit("HEAD", sig, sig, "", tree, localCommit, remoteCommit)
////
////    // Clean up
////    repo.StateCleanup()
////}