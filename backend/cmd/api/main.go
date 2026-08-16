// Command api starts the SIRAVA HTTP API.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"simrs-backend/internal/config"
	"simrs-backend/internal/modules/aktivitas_log"
	aktivitasloghttp "simrs-backend/internal/modules/aktivitas_log/delivery/http"
	"simrs-backend/internal/modules/autentikasi"
	autentikasihttp "simrs-backend/internal/modules/autentikasi/delivery/http"
	"simrs-backend/internal/modules/awal_keperawatan_igd"
	awalkeperawatanigdhttp "simrs-backend/internal/modules/awal_keperawatan_igd/delivery/http"
	"simrs-backend/internal/modules/beranda"
	berandahttp "simrs-backend/internal/modules/beranda/delivery/http"
	"simrs-backend/internal/modules/bpjs"
	bpjshttp "simrs-backend/internal/modules/bpjs/delivery/http"
	dataklaim "simrs-backend/internal/modules/bpjs/vclaim/monitoring/data_klaim"
	dataklaimhttp "simrs-backend/internal/modules/bpjs/vclaim/monitoring/data_klaim/delivery/http"
	"simrs-backend/internal/modules/cppt"
	cppthttp "simrs-backend/internal/modules/cppt/delivery/http"
	"simrs-backend/internal/modules/diagnosa_pasien"
	diagnosapasienhttp "simrs-backend/internal/modules/diagnosa_pasien/delivery/http"
	"simrs-backend/internal/modules/idrg"
	idrghttp "simrs-backend/internal/modules/idrg/delivery/http"
	"simrs-backend/internal/modules/manajemen_pengguna"
	manajemenpenggunahttp "simrs-backend/internal/modules/manajemen_pengguna/delivery/http"
	"simrs-backend/internal/modules/penanganan_dokter_petugas"
	penanganandokterpetugashttp "simrs-backend/internal/modules/penanganan_dokter_petugas/delivery/http"
	"simrs-backend/internal/modules/permintaan_radiologi"
	permintaanradiologihttp "simrs-backend/internal/modules/permintaan_radiologi/delivery/http"
	"simrs-backend/internal/modules/resume_pasien_ranap"
	resumepasienranaphttp "simrs-backend/internal/modules/resume_pasien_ranap/delivery/http"
	"simrs-backend/internal/modules/riwayat_perawatan"
	riwayatperawatanhttp "simrs-backend/internal/modules/riwayat_perawatan/delivery/http"
	"simrs-backend/internal/modules/triase_igd"
	triaseigdhttp "simrs-backend/internal/modules/triase_igd/delivery/http"
	"simrs-backend/internal/platform/database"
	faisalroutes "simrs-backend/internal/routes/faisal"
	sahrulroutes "simrs-backend/internal/routes/sahrul"
)

func main() {
	if err := config.LoadEnvFile(".env"); err != nil {
		log.Fatal(err)
	}
	databaseConfig, err := config.ApplicationDatabase()
	if err != nil {
		log.Fatal(err)
	}
	if err := config.ValidateMigrationTarget(databaseConfig); err != nil {
		log.Fatal(err)
	}
	port, err := config.ApplicationPort()
	if err != nil {
		log.Fatal(err)
	}
	tokenTTL, err := config.AuthTokenTTL()
	if err != nil {
		log.Fatal(err)
	}
	eklaimConfig, err := config.EKlaimConfig()
	if err != nil {
		log.Fatal(err)
	}
	bpjsConfig, err := config.BPJSConfig()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	db, err := database.OpenMySQL(ctx, databaseConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	simrsDatabaseConfig, err := config.SIMRSDatabase()
	if err != nil {
		log.Fatal(err)
	}
	simrsDB, err := database.OpenMySQLLazy(simrsDatabaseConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer simrsDB.Close()

	autentikasiRepositori := autentikasi.NewRepositori(db, simrsDB)
	autentikasiLayanan := autentikasi.NewLayanan(autentikasiRepositori, tokenTTL)
	autentikasiHandler := autentikasihttp.NewHandler(autentikasiLayanan)
	berandaHandler := berandahttp.NewHandler(beranda.NewRepositori(simrsDB, db))
	bpjsHandler := bpjshttp.NewHandler(bpjs.NewLayanan(bpjs.Konfigurasi{
		ConsumerID: bpjsConfig.ConsumerID,
		SecretKey:  bpjsConfig.SecretKey,
		UserKey:    bpjsConfig.UserKey,
	}))
	bpjsDataKlaimHandler := dataklaimhttp.NewHandler(dataklaim.NewLayanan(dataklaim.Konfigurasi{
		ConsumerID: bpjsConfig.ConsumerID,
		SecretKey:  bpjsConfig.SecretKey,
		UserKey:    bpjsConfig.UserKey,
		BaseURL:    bpjsConfig.VClaimURL,
	}))
	idrgRepositori := idrg.NewRepositori(simrsDB)
	idrgHandler := idrghttp.NewHandler(idrgRepositori, idrg.NewLayanan(idrgRepositori, eklaimConfig))
	manajemenPenggunaHandler := manajemenpenggunahttp.NewHandler(manajemen_pengguna.NewRepositori(db, simrsDB))
	cpptRepositori := cppt.NewRepositori(db, simrsDB)
	cpptHandler := cppthttp.NewHandler(cppt.NewLayanan(cpptRepositori))
	diagnosaPasienRepositori := diagnosa_pasien.NewRepositori(simrsDB)
	diagnosaPasienHandler := diagnosapasienhttp.NewHandler(diagnosa_pasien.NewLayanan(diagnosaPasienRepositori))
	penangananRepositori := penanganan_dokter_petugas.NewRepositori(simrsDB)
	penangananHandler := penanganandokterpetugashttp.NewHandler(penanganan_dokter_petugas.NewLayanan(penangananRepositori))
	permintaanRadiologiRepositori := permintaan_radiologi.NewRepositori(simrsDB)
	permintaanRadiologiHandler := permintaanradiologihttp.NewHandler(permintaan_radiologi.NewLayanan(permintaanRadiologiRepositori))
	riwayatPerawatanRepositori := riwayat_perawatan.NewRepositori(simrsDB, config.SIMRSWebBaseURL())
	riwayatPerawatanHandler := riwayatperawatanhttp.NewHandler(riwayat_perawatan.NewLayanan(riwayatPerawatanRepositori))
	triaseIGDRepositori := triase_igd.NewRepositori(simrsDB)
	triaseIGDHandler := triaseigdhttp.NewHandler(triase_igd.NewLayanan(triaseIGDRepositori))
	awalKeperawatanIGDRepositori := awal_keperawatan_igd.NewRepositori(simrsDB)
	awalKeperawatanIGDHandler := awalkeperawatanigdhttp.NewHandler(awal_keperawatan_igd.NewLayanan(awalKeperawatanIGDRepositori))
	resumePasienRanapRepositori := resume_pasien_ranap.NewRepositori(simrsDB)
	resumePasienRanapHandler := resumepasienranaphttp.NewHandler(resume_pasien_ranap.NewLayanan(resumePasienRanapRepositori))
	aktivitasLogRepositori := aktivitas_log.NewRepositori(db, simrsDB)
	aktivitasLogHandler := aktivitasloghttp.NewHandler(aktivitasLogRepositori)

	router := gin.New()
	router.Use(gin.Logger(), aktivitas_log.Middleware(aktivitasLogRepositori), gin.Recovery())
	faisalroutes.Register(router, faisalroutes.Dependencies{
		AktivitasLog:            aktivitasLogHandler,
		Autentikasi:             autentikasiHandler,
		Beranda:                 berandaHandler,
		BPJS:                    bpjsHandler,
		BPJSDataKlaim:           bpjsDataKlaimHandler,
		IDRG:                    idrgHandler,
		ManajemenPengguna:       manajemenPenggunaHandler,
		CPPT:                    cpptHandler,
		DiagnosaPasien:          diagnosaPasienHandler,
		PenangananDokterPetugas: penangananHandler,
		PermintaanRadiologi:     permintaanRadiologiHandler,
		RiwayatPerawatan:        riwayatPerawatanHandler,
		TriaseIGD:               triaseIGDHandler,
		AwalKeperawatanIGD:      awalKeperawatanIGDHandler,
		ResumePasienRanap:       resumePasienRanapHandler,
	})
	sahrulroutes.Register(router, sahrulroutes.Dependencies{})

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("SIRAVA API listening on http://localhost:%d", port)
		serverErrors <- server.ListenAndServe()
	}()

	shutdown, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	select {
	case <-shutdown.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			log.Printf("server shutdown error: %v", err)
		}
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}
}
