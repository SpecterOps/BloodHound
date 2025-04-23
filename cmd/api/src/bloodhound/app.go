package bloodhound

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"

	"github.com/specterops/bloodhound/bomenc"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/util"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/model"
)

type App struct {
	db      database.Database
	graphDB graph.Database
	cfg     config.Configuration

	// Any number of exported services that are containers to pin methods on
	// without polluting app into another nasty mess
	IngestService IngestService
}

func NewApp(db database.Database, graphDB graph.Database, cfg config.Configuration) App {
	app := App{
		db:      db,
		graphDB: graphDB,
		cfg:     cfg,

		IngestService: IngestService{db: db, graphDB: graphDB, cfg: cfg},
	}

	// Initialize each service with the necessary dependancies
	// Todo: Determine if the ingest service needs all of these?
	return app
}

// Each service can have all of its logic contained, while having access to the
// dependencies it needs.

// TODO: Put this somewhere else that isn't app.go
type IngestService struct {
	db      database.Database
	graphDB graph.Database
	cfg     config.Configuration
}

func (s IngestService) ProcessIngestTasks(ctx context.Context) error {
	if ingestTasks, err := s.db.GetAllIngestTasks(ctx); err != nil {
		return fmt.Errorf("get all ingest tasks: %v", err)
	} else if err := s.db.SetDatapipeStatus(ctx, model.DatapipeStatusIngesting, false); err != nil {
		return fmt.Errorf("set datapipe status ingesting: %v", err)
	} else {
		var errs = make([]error, 0)

		for _, ingestTask := range ingestTasks {
			ingestTaskLogger := slog.Default().With(
				slog.Group("ingest_task",
					slog.Int64("id", ingestTask.ID),
					slog.String("file_name", ingestTask.FileName),
				),
			)

			// Check the context to see if we should continue processing ingest tasks. This has to be explicit since error
			// handling assumes that all failures should be logged and not returned.
			if ctx.Err() != nil {
				errs = append(errs, fmt.Errorf("context error encountered: %v", err))
				return errors.Join(errs...)
			}

			if paths, failed, err := preProcessIngestFile(ctx, s.cfg.TempDirectory(), ingestTask); errors.Is(err, fs.ErrNotExist) {
				ingestTaskLogger.WarnContext(
					ctx,
					"File does not exist for ingest task",
					slog.String("err", err.Error()),
				)
			} else if err != nil {
				ingestTaskLogger.ErrorContext(
					ctx,
					"Failed to preprocess ingest file",
					slog.String("err", err.Error()),
				)
				errs = append(errs, fmt.Errorf("preprocess ingest file: %v", err))
			} else if total, failed, err := processIngestFile(ctx, s.graphDB, paths, failed); err != nil {
				ingestTaskLogger.ErrorContext(
					ctx,
					"Failed to process ingest file",
					slog.String("err", err.Error()),
				)
				errs = append(errs, fmt.Errorf("process ingest file: %v", err))
			} else if job, err := s.db.GetIngestJob(ctx, ingestTask.TaskID); err != nil {
				ingestTaskLogger.ErrorContext(
					ctx,
					"Failed to get ingest job",
					slog.String("err", err.Error()),
				)
				errs = append(errs, fmt.Errorf("get ingest job: %v", err))
			} else if err := updateIngestJob(ctx, s.db, job, total, failed); err != nil {
				ingestTaskLogger.ErrorContext(
					ctx,
					"Failed to update file completion for ingest job",
					slog.String("err", err.Error()),
				)
				errs = append(errs, fmt.Errorf("update ingest job: %v", err))
			}

			if err := s.db.DeleteIngestTask(ctx, ingestTask); err != nil {
				ingestTaskLogger.ErrorContext(
					ctx,
					"Failed to remove ingest task",
					slog.String("err", err.Error()),
				)
				errs = append(errs, fmt.Errorf("delete ingest task: %v", err))
			}
		}
		return errors.Join(errs...)
	}
}

func updateIngestJob(ctx context.Context, db database.Database, job model.IngestJob, total int, failed int) error {
	job.TotalFiles = total
	job.FailedFiles += failed

	if err := db.UpdateIngestJob(ctx, job); err != nil {
		return fmt.Errorf("could not update file completion for ingest job id %d: %w", job.ID, err)
	} else {
		return nil
	}
}

func preProcessIngestFile(ctx context.Context, tmpDir string, ingestTask model.IngestTask) ([]string, int, error) {
	if ingestTask.FileType == model.FileTypeJson {
		//If this isn't a zip file, just return a slice with the path in it and let stuff process as normal
		return []string{ingestTask.FileName}, 0, nil
	} else if archive, err := zip.OpenReader(ingestTask.FileName); err != nil {
		return []string{}, 0, err
	} else {
		var (
			errs      = util.NewErrorCollector()
			failed    = 0
			filePaths = make([]string, len(archive.File))
		)

		for i, f := range archive.File {
			//skip directories
			if f.FileInfo().IsDir() {
				continue
			}
			// Break out if temp file creation fails
			// Collect errors for other failures within the archive
			if tempFile, err := os.CreateTemp(tmpDir, "bh"); err != nil {
				return []string{}, 0, err
			} else if srcFile, err := f.Open(); err != nil {
				errs.Add(fmt.Errorf("error opening file %s in archive %s: %v", f.Name, ingestTask.FileName, err))
				failed++
			} else if normFile, err := bomenc.NormalizeToUTF8(srcFile); err != nil {
				errs.Add(fmt.Errorf("error normalizing file %s to UTF8 in archive %s: %v", f.Name, ingestTask.FileName, err))
				failed++
			} else if _, err := io.Copy(tempFile, normFile); err != nil {
				errs.Add(fmt.Errorf("error extracting file %s in archive %s: %v", f.Name, ingestTask.FileName, err))
				failed++
			} else if err := tempFile.Close(); err != nil {
				errs.Add(fmt.Errorf("error closing temp file %s: %v", f.Name, err))
				failed++
			} else {
				filePaths[i] = tempFile.Name()
			}
		}

		//Close the archive and delete it
		if err := archive.Close(); err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Error closing archive %s: %v", ingestTask.FileName, err))
		} else if err := os.Remove(ingestTask.FileName); err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Error deleting archive %s: %v", ingestTask.FileName, err))
		}

		return filePaths, failed, errs.Combined()
	}
}

func processIngestFile(ctx context.Context, graphDB graph.Database, paths []string, failed int) (int, int, error) {
	return len(paths), failed, graphDB.BatchOperation(ctx, func(batch graph.Batch) error {
		for _, filePath := range paths {
			file, err := os.Open(filePath)
			if err != nil {
				failed++
				return err
			} else if err := ReadFileForIngest(batch, file); err != nil {
				failed++
				slog.ErrorContext(ctx, fmt.Sprintf("Error reading ingest file %s: %v", filePath, err))
			}

			if err := file.Close(); err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("Error closing ingest file %s: %v", filePath, err))
			} else if err := os.Remove(filePath); errors.Is(err, fs.ErrNotExist) {
				slog.WarnContext(ctx, fmt.Sprintf("Removing ingest file %s: %v", filePath, err))
			} else if err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("Error removing ingest file %s: %v", filePath, err))
			}
		}

		return nil
	})
}
