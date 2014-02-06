package bundlecollection_test

import (
	. "bosh/agent/applier/bundlecollection"
	fakesys "bosh/system/fakes"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
)

var _ = Describe("FileBundle", func() {
	var (
		fs          *fakesys.FakeFileSystem
		installPath string
		enablePath  string
		fileBundle  FileBundle
	)

	BeforeEach(func() {
		fs = &fakesys.FakeFileSystem{}
		installPath = "/install-path"
		enablePath = "/enable-path"
	})

	JustBeforeEach(func() {
		fileBundle = NewFileBundle(installPath, enablePath, fs)
	})

	Describe("#Install", func() {
		It("Installs the bundle at the given path with the correct permissions", func() {
			actualFs, path, err := fileBundle.Install()

			Expect(err).NotTo(HaveOccurred())
			Expect(actualFs).To(Equal(fs))
			Expect(path).To(Equal(installPath))
			Expect(actualFs.FileExists(installPath)).To(BeTrue())
			fileStats := fs.GetFileTestStat(installPath)
			Expect(fileStats).ToNot(BeNil())
			Expect(fileStats.FileType).To(Equal(fakesys.FakeFileType(fakesys.FakeFileTypeDir)))
			Expect(os.FileMode(0755)).To(Equal(fileStats.FileMode))
		})

		It("Errors when bundle cannot be installed", func() {
			fs.MkdirAllError = errors.New("fake-mkdirall-error")

			_, _, err := fileBundle.Install()

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-mkdirall-error"))
		})
	})

	Describe("#GetInstallPath", func() {
		It("Returns the install path", func() {
			fs.MkdirAll(installPath, os.FileMode(0))

			actualFs, installPath, err := fileBundle.GetInstallPath()

			Expect(err).NotTo(HaveOccurred())
			Expect(actualFs).To(Equal(fs))
			Expect(installPath).To(Equal(installPath))
		})

		It("Errors when the directory does not exist", func() {
			_, _, err := fileBundle.GetInstallPath()

			Expect(err).To(HaveOccurred())
		})
	})

	Describe("#Enable", func() {
		Context("when bundle is install", func() {
			It("Succeeds", func() {
				_, _, err := fileBundle.Install()
				Expect(err).NotTo(HaveOccurred())

				err = fileBundle.Enable()
				Expect(err).NotTo(HaveOccurred())

				// symlink exists
				fileStats := fs.GetFileTestStat(enablePath)
				Expect(fileStats).NotTo(BeNil())
				Expect(fileStats.FileType).To(Equal(fakesys.FakeFileType(fakesys.FakeFileTypeSymlink)))
				Expect(installPath).To(Equal(fileStats.SymlinkTarget))

				// enable directory is created
				fileStats = fs.GetFileTestStat("/")
				Expect(fileStats).NotTo(BeNil())
				Expect(fileStats.FileType).To(Equal(fakesys.FakeFileType(fakesys.FakeFileTypeDir)))
				Expect(fileStats.FileMode).To(Equal(os.FileMode(0755)))

				// check idempotency
				err = fileBundle.Enable()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when bundle is not install", func() {
			It("errors", func() {
				err := fileBundle.Enable()
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(Equal("bundle must be installed"))

				fileStats := fs.GetFileTestStat(enablePath)
				Expect(fileStats).To(BeNil())
			})
		})

		Context("when enable dir cannot be created", func() {
			It("errors", func() {
				_, _, err := fileBundle.Install()
				Expect(err).NotTo(HaveOccurred())
				fs.MkdirAllError = errors.New("fake-mkdirall-error")

				err = fileBundle.Enable()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-mkdirall-error"))
			})
		})

		Context("when bundle cannot be enabled", func() {
			It("errors", func() {
				_, _, err := fileBundle.Install()
				Expect(err).NotTo(HaveOccurred())
				fs.SymlinkError = errors.New("fake-symlink-error")

				err = fileBundle.Enable()

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-symlink-error"))
			})
		})
	})
})
