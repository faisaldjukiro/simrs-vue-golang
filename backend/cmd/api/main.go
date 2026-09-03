// Command api starts the SIRAPI HTTP API.
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
	"simrs-backend/internal/modules/awal_medis_igd"
	awalMedisIgdHttp "simrs-backend/internal/modules/awal_medis_igd/delivery/http"
	"simrs-backend/internal/modules/awal_medis_ranap"
	awalmedisranaphttp "simrs-backend/internal/modules/awal_medis_ranap/delivery/http"
	"simrs-backend/internal/modules/awal_medis_umum"
	awalmedisumumhttp "simrs-backend/internal/modules/awal_medis_umum/delivery/http"
	"simrs-backend/internal/modules/beranda"
	berandahttp "simrs-backend/internal/modules/beranda/delivery/http"
	"simrs-backend/internal/modules/berkas_digital"
	berkasdigitalhttp "simrs-backend/internal/modules/berkas_digital/delivery/http"
	"simrs-backend/internal/modules/bpjs"
	bpjshttp "simrs-backend/internal/modules/bpjs/delivery/http"
	dataklaim "simrs-backend/internal/modules/bpjs/vclaim/monitoring/data_klaim"
	dataklaimhttp "simrs-backend/internal/modules/bpjs/vclaim/monitoring/data_klaim/delivery/http"
	"simrs-backend/internal/modules/cppt"
	cppthttp "simrs-backend/internal/modules/cppt/delivery/http"
	"simrs-backend/internal/modules/diagnosa_pasien"
	diagnosapasienhttp "simrs-backend/internal/modules/diagnosa_pasien/delivery/http"
	"simrs-backend/internal/modules/ews_ranap"
	ewsranaphttp "simrs-backend/internal/modules/ews_ranap/delivery/http"
	"simrs-backend/internal/modules/idrg"
	idrghttp "simrs-backend/internal/modules/idrg/delivery/http"
	"simrs-backend/internal/modules/implementasi_keperawatan"
	implementasikeperawatanhttp "simrs-backend/internal/modules/implementasi_keperawatan/delivery/http"
	"simrs-backend/internal/modules/kelola_menu"
	kelolamenuhttp "simrs-backend/internal/modules/kelola_menu/delivery/http"
	"simrs-backend/internal/modules/laporan_10_penyakit"
	laporan10penyakithttp "simrs-backend/internal/modules/laporan_10_penyakit/delivery/http"
	"simrs-backend/internal/modules/laporan_kunjungan_ralan"
	laporankunjunganralanhttp "simrs-backend/internal/modules/laporan_kunjungan_ralan/delivery/http"
	"simrs-backend/internal/modules/laporan_kunjungan_ranap"
	laporankunjunganranaphttp "simrs-backend/internal/modules/laporan_kunjungan_ranap/delivery/http"
	"simrs-backend/internal/modules/manajemen_pengguna"
	manajemenpenggunahttp "simrs-backend/internal/modules/manajemen_pengguna/delivery/http"
	"simrs-backend/internal/modules/penanganan_dokter_petugas"
	penanganandokterpetugashttp "simrs-backend/internal/modules/penanganan_dokter_petugas/delivery/http"
	"simrs-backend/internal/modules/permintaan_laboratorium"
	permintaanlaboratoriumhttp "simrs-backend/internal/modules/permintaan_laboratorium/delivery/http"
	"simrs-backend/internal/modules/permintaan_radiologi"
	permintaanradiologihttp "simrs-backend/internal/modules/permintaan_radiologi/delivery/http"
	"simrs-backend/internal/modules/resep"
	resephttp "simrs-backend/internal/modules/resep/delivery/http"
	"simrs-backend/internal/modules/resume_pasien"
	resumepasienhttp "simrs-backend/internal/modules/resume_pasien/delivery/http"
	"simrs-backend/internal/modules/resume_pasien_ranap"
	resumepasienranaphttp "simrs-backend/internal/modules/resume_pasien_ranap/delivery/http"
	"simrs-backend/internal/modules/riwayat_perawatan"
	riwayatperawatanhttp "simrs-backend/internal/modules/riwayat_perawatan/delivery/http"
	"simrs-backend/internal/modules/triase_igd"
	triaseigdhttp "simrs-backend/internal/modules/triase_igd/delivery/http"
	"simrs-backend/internal/modules/ventilator"
	ventilatorhttp "simrs-backend/internal/modules/ventilator/delivery/http"
	"simrs-backend/internal/modules/whatsapp_gateway"
	whatsappgatewayhttp "simrs-backend/internal/modules/whatsapp_gateway/delivery/http"
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
	whatsappConfig := config.WhatsAppConfig()

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
	berandaHandler := berandahttp.NewHandler(
		beranda.NewRepositori(simrsDB, db),
		beranda.NewPemeriksaEKlaim(eklaimConfig.WSURL, eklaimConfig.Key),
		beranda.NewPemeriksaBPJS(bpjsConfig.VClaimURL, bpjsConfig.ConsumerID, bpjsConfig.SecretKey, bpjsConfig.UserKey),
	)
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
	kelolaMenuHandler := kelolamenuhttp.NewHandler(kelola_menu.NewRepositori(db))
	laporan10PenyakitHandler := laporan10penyakithttp.NewHandler(
		laporan_10_penyakit.NewRepositori(simrsDB),
	)
	laporanKunjunganRalanHandler := laporankunjunganralanhttp.NewHandler(laporan_kunjungan_ralan.NewRepositori(simrsDB))
	laporanKunjunganRanapHandler := laporankunjunganranaphttp.NewHandler(laporan_kunjungan_ranap.NewRepositori(simrsDB))
	manajemenPenggunaHandler := manajemenpenggunahttp.NewHandler(manajemen_pengguna.NewRepositori(db, simrsDB))
	cpptRepositori := cppt.NewRepositori(db, simrsDB)
	cpptHandler := cppthttp.NewHandler(cppt.NewLayanan(cpptRepositori))
	diagnosaPasienRepositori := diagnosa_pasien.NewRepositori(simrsDB)
	diagnosaPasienHandler := diagnosapasienhttp.NewHandler(diagnosa_pasien.NewLayanan(diagnosaPasienRepositori))
	ewsRanapRepositori := ews_ranap.NewRepositori(db, simrsDB)
	ewsRanapHandler := ewsranaphttp.NewHandler(ews_ranap.NewLayanan(ewsRanapRepositori))
	implementasiKeperawatanRepositori := implementasi_keperawatan.NewRepositori(db, simrsDB)
	implementasiKeperawatanHandler := implementasikeperawatanhttp.NewHandler(implementasi_keperawatan.NewLayanan(implementasiKeperawatanRepositori))
	penangananRepositori := penanganan_dokter_petugas.NewRepositori(simrsDB)
	penangananHandler := penanganandokterpetugashttp.NewHandler(penanganan_dokter_petugas.NewLayanan(penangananRepositori))
	permintaanRadiologiRepositori := permintaan_radiologi.NewRepositori(simrsDB)
	permintaanRadiologiHandler := permintaanradiologihttp.NewHandler(permintaan_radiologi.NewLayanan(permintaanRadiologiRepositori))
	permintaanLaboratoriumRepositori := permintaan_laboratorium.NewRepositori(simrsDB)
	permintaanLaboratoriumHandler := permintaanlaboratoriumhttp.NewHandler(permintaan_laboratorium.NewLayanan(permintaanLaboratoriumRepositori))
	berkasDigitalRepositori := berkas_digital.NewRepositori(simrsDB, config.SIMRSWebBaseURL())
	berkasDigitalHandler := berkasdigitalhttp.NewHandler(
		berkas_digital.NewLayanan(berkasDigitalRepositori),
		config.BerkasDigitalUploadURL(),
		config.BerkasDigitalLokasiPrefix(),
	)
	riwayatPerawatanRepositori := riwayat_perawatan.NewRepositori(simrsDB, config.SIMRSWebBaseURL())
	riwayatPerawatanHandler := riwayatperawatanhttp.NewHandler(riwayat_perawatan.NewLayanan(riwayatPerawatanRepositori))
	triaseIGDRepositori := triase_igd.NewRepositori(simrsDB)
	triaseIGDHandler := triaseigdhttp.NewHandler(triase_igd.NewLayanan(triaseIGDRepositori))
	awalKeperawatanIGDRepositori := awal_keperawatan_igd.NewRepositori(simrsDB)
	awalKeperawatanIGDHandler := awalkeperawatanigdhttp.NewHandler(awal_keperawatan_igd.NewLayanan(awalKeperawatanIGDRepositori))
	awalMedisUmumRepositori := awal_medis_umum.NewRepositori(simrsDB)
	awalMedisUmumHandler := awalmedisumumhttp.NewHandler(awal_medis_umum.NewLayanan(awalMedisUmumRepositori))
	awalMedisRanapRepositori := awal_medis_ranap.NewRepositori(simrsDB)
	awalMedisIgdRepositori := awal_medis_igd.NewRepositori(simrsDB)
	awalMedisRanapHandler := awalmedisranaphttp.NewHandler(awal_medis_ranap.NewLayanan(awalMedisRanapRepositori))
	awalMedisIgdHandler := awalMedisIgdHttp.NewHandler(awal_medis_igd.NewLayanan(awalMedisIgdRepositori))
	resumePasienRanapRepositori := resume_pasien_ranap.NewRepositori(simrsDB)
	resumePasienRanapHandler := resumepasienranaphttp.NewHandler(resume_pasien_ranap.NewLayanan(resumePasienRanapRepositori))
	resumePasienRepositori := resume_pasien.NewRepositori(simrsDB)
	resumePasienHandler := resumepasienhttp.NewHandler(resume_pasien.NewLayanan(resumePasienRepositori))
	resepRepositori := resep.NewRepositori(simrsDB, simrsDB)
	resepHandler := resephttp.NewHandler(resep.NewLayanan(resepRepositori))
	aktivitasLogRepositori := aktivitas_log.NewRepositori(db, simrsDB)
	aktivitasLogHandler := aktivitasloghttp.NewHandler(aktivitasLogRepositori)
	whatsappGatewayHandler := whatsappgatewayhttp.NewHandler(whatsapp_gateway.NewLayanan(
		whatsappConfig.URL,
		whatsappConfig.Key,
		whatsappConfig.Timeout,
	))
	ventilatorHandler := ventilatorhttp.NewHandler(ventilator.NewRepositori(simrsDB))

	router := gin.New()
	router.Use(gin.Logger(), aktivitas_log.Middleware(aktivitasLogRepositori), gin.Recovery())
	faisalroutes.Register(router, faisalroutes.Dependencies{
		AktivitasLog:            aktivitasLogHandler,
		Autentikasi:             autentikasiHandler,
		Beranda:                 berandaHandler,
		BPJS:                    bpjsHandler,
		BPJSDataKlaim:           bpjsDataKlaimHandler,
		IDRG:                    idrgHandler,
		KelolaMenu:              kelolaMenuHandler,
		Laporan10Penyakit:       laporan10PenyakitHandler,
		LaporanKunjunganRalan:   laporanKunjunganRalanHandler,
		LaporanKunjunganRanap:   laporanKunjunganRanapHandler,
		ManajemenPengguna:       manajemenPenggunaHandler,
		CPPT:                    cpptHandler,
		DiagnosaPasien:          diagnosaPasienHandler,
		EWSRanap:                ewsRanapHandler,
		ImplementasiKeperawatan: implementasiKeperawatanHandler,
		PenangananDokterPetugas: penangananHandler,
		PermintaanRadiologi:     permintaanRadiologiHandler,
		PermintaanLaboratorium:  permintaanLaboratoriumHandler,
		BerkasDigital:           berkasDigitalHandler,
		RiwayatPerawatan:        riwayatPerawatanHandler,
		TriaseIGD:               triaseIGDHandler,
		AwalKeperawatanIGD:      awalKeperawatanIGDHandler,
		AwalMedisUmum:           awalMedisUmumHandler,
		AwalMedisRanap:          awalMedisRanapHandler,
		AwalMedisIgd:            awalMedisIgdHandler,
		ResumePasien:            resumePasienHandler,
		ResumePasienRanap:       resumePasienRanapHandler,
		Resep:                   resepHandler,
		WhatsAppGateway:         whatsappGatewayHandler,
		Ventilator:              ventilatorHandler,
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
		log.Printf("SIRAPI API listening on http://localhost:%d", port)
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
