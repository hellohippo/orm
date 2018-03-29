package integration_test

import (
	"io/ioutil"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Migration Setup", func() {
	var cmd *exec.Cmd

	BeforeEach(func() {
		dir, err := ioutil.TempDir("", "gom")
		Expect(err).To(BeNil())

		cmd = exec.Command(gomPath, "--database-url", "sqlite3://gom.db", "migration", "setup")
		cmd.Dir = dir
	})

	It("setups the project successfully", func() {
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))
		Eventually(session.Err).Should(gbytes.Say("The project has been configured successfully"))

		path := filepath.Join(cmd.Dir, "/database/migration/00060524000000_setup.sql")
		Expect(path).To(BeARegularFile())
	})

	Context("when the setup command is executed more than once", func() {
		It("returns an error", func() {
			setupCmd := exec.Command(gomPath, "--database-url", "sqlite3://gom.db", "migration", "setup")
			setupCmd.Dir = cmd.Dir

			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			session, err = gexec.Start(setupCmd, GinkgoWriter, GinkgoWriter)
			Eventually(session).Should(gexec.Exit(103))

			Expect(session.Err).To(gbytes.Say("/database/migration/00060524000000_setup.sql' already exists"))
		})
	})

	Context("when the database is not available", func() {
		BeforeEach(func() {
			cmd.Args = []string{gomPath, "migration", "setup"}
		})

		It("returns an error", func() {
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(101))
			Expect(session.Err).To(gbytes.Say(`sql: unknown driver`))
		})
	})
})