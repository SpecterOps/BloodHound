package graphify

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/services/storage"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bomenc"
	"github.com/specterops/dawgs/util"
)

func SpoolToScratch(ctx context.Context, tempDirectory string, fileService storage.FileService, storedFileName string) (string, error) {
	var (
		sourceFile  io.ReadCloser
		scratchFile *os.File
		scratchPath string
		err         error
		success     bool
	)

	sourceFile, _, err = fileService.GetFile(ctx, storedFileName)
	if err != nil {
		return "", fmt.Errorf("open stored ingest file %q: %w", storedFileName, err)
	}
	defer sourceFile.Close()

	scratchFile, err = os.CreateTemp(tempDirectory, "ingest-archive-*")
	if err != nil {
		return "", fmt.Errorf("create ingest scratch file: %w", err)
	}

	scratchPath = scratchFile.Name()

	defer func() {
		if scratchFile != nil {
			_ = scratchFile.Close()
		}
		if !success {
			_ = os.Remove(scratchPath)
		}
	}()

	if _, err = io.Copy(scratchFile, sourceFile); err != nil {
		return "", fmt.Errorf("copy stored ingest file %q to scratch: %w", storedFileName, err)
	}

	if err = scratchFile.Close(); err != nil {
		return "", fmt.Errorf("close ingest scratch file %q: %w", scratchPath, err)
	}

	scratchFile = nil
	success = true

	return scratchPath, nil
}

func OpenScratchReadSeeker(ctx context.Context, tempDirectory string, fileService storage.FileService, storedFileName string) (*os.File, string, error) {
	var (
		scratchPath string
		scratchFile *os.File
		err         error
	)

	if scratchPath, err = SpoolToScratch(ctx, tempDirectory, fileService, storedFileName); err != nil {
		return nil, "", err
	}

	if scratchFile, err = os.Open(scratchPath); err != nil {
		_ = os.Remove(scratchPath)
		return nil, "", fmt.Errorf("error opening ingest scratch file %q: %w", scratchPath, err)
	}

	return scratchFile, scratchPath, nil
}

func WriteArchiveFileToStorage(ctx context.Context, fileService storage.FileService, archiveFile *zip.File, prefix string) (string, error) {
	var (
		sourceFile     io.ReadCloser
		normalizedFile io.Reader
		err            error
	)

	if sourceFile, err = archiveFile.Open(); err != nil {
		return "", fmt.Errorf("error opening archive file %q: %w", archiveFile.Name, err)
	}
	defer sourceFile.Close()

	if normalizedFile, err = bomenc.NormalizeToUTF8(sourceFile); err != nil {
		return "", fmt.Errorf("error normalizing archive file %q to UTF8: %w", archiveFile.Name, err)
	}

	extractedPath, err := fileService.WriteTempFile(
		ctx,
		prefix,
		normalizedFile,
		storage.WriteOptions{}, // TODO MC: include write options?
	)
	if err != nil {
		return "", fmt.Errorf("write archive file %q to stroage: %w", archiveFile.Name, err)
	}

	return extractedPath, nil
}

func ExtractIngestFiles(ctx context.Context, tempDirectory string, fileService storage.FileService, storedFileName, providedFileName string, fileType model.FileType, prefix string) ([]IngestFileData, error) {
	if fileType == model.FileTypeJson {
		// If this isn't a zip file, just return a slice with the path in it and let stuff process as normal
		return []IngestFileData{
			{
				Name: providedFileName,
				Path: storedFileName,
			},
		}, nil
	}

	// Zip Path:
	scratchPath, err := SpoolToScratch(ctx, tempDirectory, fileService, storedFileName)
	if err != nil {
		return []IngestFileData{
			{
				Name:   providedFileName,
				Path:   storedFileName,
				Errors: []string{fmt.Sprintf("Error spooling archive to scratch: %v", err)},
			},
		}, err
	}
	defer os.Remove(scratchPath)

	archive, err := zip.OpenReader(scratchPath)
	if err != nil {
		return []IngestFileData{
			{
				Name:   providedFileName,
				Path:   storedFileName,
				Errors: []string{fmt.Sprintf("Error opening archive: %v", err)},
			},
		}, err
	}
	defer archive.Close()

	var (
		errs     = util.NewErrorCollector()
		fileData = make([]IngestFileData, 0)
	)
	for _, archiveFile := range archive.File {
		if archiveFile.FileInfo().IsDir() {
			continue
		}

		processedFileData := IngestFileData{
			Name:       archiveFile.Name,
			ParentFile: providedFileName,
		}

		//if extractedPath, err := WriteArchiveFileToStorage(ctx, fileService, archiveFile, fmt.Sprintf("tmp/file_upload_job%d_", jobID)); err != nil {
		if extractedPath, err := WriteArchiveFileToStorage(ctx, fileService, archiveFile, prefix); err != nil {
			processedFileData.Errors = []string{err.Error()}

			errs.Add(fmt.Errorf(
				"error extracting file %s in archive %s: %w",
				archiveFile.Name,
				storedFileName,
				err,
			))
		} else {
			processedFileData.Path = extractedPath
		}

		fileData = append(fileData, processedFileData)
	}

	if err := fileService.DeleteFile(ctx, storedFileName); err != nil {
		slog.ErrorContext(
			ctx,
			"Error deleting archive",
			slog.String("path", storedFileName),
			attr.Error(err),
		)
	}

	return fileData, errs.Combined()
}
