package rest

import (
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"

	"github.com/couchbaselabs/sync_gateway/base"
	"github.com/couchbaselabs/sync_gateway/db"
)

// HTTP handler for GET _design/$ddoc
func (h *handler) handleGetDesignDoc() error {
	ddocID := h.PathVar("ddoc")
	base.TEMP("GetDesignDoc %q", ddocID)
	var result interface{}
	if ddocID == "sync_gateway" {
		// we serve this content here so that CouchDB 1.2 has something to
		// hash into the replication-id, to correspond to our filter.
		filter := "ok"
		if h.db.DatabaseContext.ChannelMapper != nil {
			hash := sha1.New()
			io.WriteString(hash, h.db.DatabaseContext.ChannelMapper.Function())
			filter = fmt.Sprint(hash.Sum(nil))
		}
		result = db.Body{"filters": db.Body{"bychannel": filter}}
	} else {
		if err := h.db.GetDesignDoc(ddocID, &result); err != nil {
			return err
		}
	}
	h.writeJSON(result)
	return nil
}

// HTTP handler for PUT _design/$ddoc
func (h *handler) handlePutDesignDoc() error {
	ddocID := h.PathVar("ddoc")
	var ddoc db.DesignDoc
	err := h.readJSONInto(&ddoc)
	if err != nil {
		return err
	}
	if err = h.db.PutDesignDoc(ddocID, ddoc); err != nil {
		return err
	}
	h.writeStatus(http.StatusCreated, "OK")
	return nil
}

// HTTP handler for DELETE _design/$ddoc
func (h *handler) handleDeleteDesignDoc() error {
	ddocID := h.PathVar("ddoc")
	return h.db.DeleteDesignDoc(ddocID)
}

// HTTP handler for GET _design/$ddoc/_view/$view
func (h *handler) handleView() error {
	ddocName := h.PathVar("ddoc")
	if ddocName == "" {
		ddocName = "sync_gateway"
	}
	viewName := h.PathVar("view")
	opts := db.Body{}
	qStale := h.getQuery("stale")
	if "" != qStale {
		opts["stale"] = qStale == "true"
	}
	qReduce := h.getQuery("reduce")
	if "" != qReduce {
		opts["reduce"] = qReduce == "true"
	}
	opts["inclusive_end"] = h.getOptBoolQuery("inclusive_end", true)
	var err error
	if key, err := h.getJSONQuery("startkey"); err != nil {
		return err
	} else if key != nil {
		opts["startkey"] = key
	}
	if key, err := h.getJSONQuery("endkey"); err != nil {
		return err
	} else if key != nil {
		opts["endkey"] = key
	}
	qGroupLevel := h.getQuery("group_level")
	if "" != qGroupLevel {
		opts["group_level"] = int(h.getIntQuery("group_level", 1))
	}
	opts["group"] = h.getOptBoolQuery("group", false)
	qLimit := h.getQuery("limit")
	if "" != qLimit {
		opts["limit"] = int(h.getIntQuery("limit", 1))
	}
	base.LogTo("HTTP", "JSON view %q/%q - opts %v", ddocName, viewName, opts)

	result, err := h.db.QueryDesignDoc(ddocName, viewName, opts)
	if err != nil {
		return err
	}
	h.setHeader("Content-Type", `application/json; charset="UTF-8"`)
	h.writeJSON(result)
	return nil
}