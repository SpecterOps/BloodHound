package samlidp

import (
	"bytes"
	"compress/flate"
	"context"
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"text/template"
	"time"

	"github.com/specterops/bloodhound/src/auth/bhsaml"
	"github.com/beevik/etree"
	"github.com/crewjam/saml"
	"github.com/crewjam/saml/logger"
	"github.com/crewjam/saml/xmlenc"
	xrv "github.com/mattermost/xml-roundtrip-validator"
	dsig "github.com/russellhaering/goxmldsig"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/mediatypes"
)

// Session represents a user session. It is returned by the
// SessionProvider implementation's GetSession method. Fields here
// are used to set fields in the SAML assertion.
type Session struct {
	ID         string
	CreateTime time.Time
	ExpireTime time.Time
	Index      string

	NameID                string
	Groups                []string
	UserName              string
	UserEmail             string
	UserCommonName        string
	UserSurname           string
	UserGivenName         string
	UserScopedAffiliation string

	CustomAttributes []saml.Attribute
}

// SessionProvider is an interface used by IdentityProvider to determine the
// Session associated with a request. For an example implementation, see
// GetSession in the samlidp package.
type SessionProvider interface {
	// GetSession returns the remote user session associated with the http.Request.
	//
	// If (and only if) the request is not associated with a session then GetSession
	// must complete the HTTP request and return nil.
	GetSession(w http.ResponseWriter, r *http.Request, req *IdpAuthnRequest) *Session
}

// ServiceProviderProvider is an interface used by IdentityProvider to look up
// service provider metadata for a request.
type ServiceProviderProvider interface {
	// GetServiceProvider returns the Service Provider metadata for the
	// service provider ID, which is typically the service provider's
	// metadata URL. If an appropriate service provider cannot be found then
	// the returned error must be os.ErrNotExist.
	GetServiceProvider(r *http.Request, serviceProviderID string) (*saml.EntityDescriptor, error)
}

// AssertionMaker is an interface used by IdentityProvider to construct the
// assertion for a request. The default implementation is DefaultAssertionMaker,
// which is used if not AssertionMaker is specified.
type AssertionMaker interface {
	// MakeAssertion constructs an assertion from session and the request and
	// assigns it to req.Assertion.
	MakeAssertion(req *IdpAuthnRequest, session *Session) error
}

// IdentityProvider implements the SAML Identity Provider role (IDP).
//
// An identity provider receives SAML assertion requests and responds
// with SAML Assertions.
//
// You must provide a keypair that is used to
// sign assertions.
//
// You must provide an implementation of ServiceProviderProvider which
// returns
//
// You must provide an implementation of the SessionProvider which
// handles the actual authentication (i.e. prompting for a username
// and password).
type IdentityProvider struct {
	Key                     crypto.PrivateKey
	Logger                  logger.Interface
	Certificate             *x509.Certificate
	Intermediates           []*x509.Certificate
	AllowUntrustedSP        bool
	MetadataURL             url.URL
	SSOURL                  url.URL
	LogoutURL               url.URL
	ServiceProviderProvider ServiceProviderProvider
	SessionProvider         SessionProvider
	AssertionMaker          AssertionMaker
	SignatureMethod         string
	ValidDuration           *time.Duration
}

// Metadata returns the metadata structure for this identity provider.
func (s *IdentityProvider) Metadata() *saml.EntityDescriptor {
	certStr := base64.StdEncoding.EncodeToString(s.Certificate.Raw)

	var validDuration time.Duration
	if s.ValidDuration != nil {
		validDuration = *s.ValidDuration
	} else {
		validDuration = saml.DefaultValidDuration
	}

	ed := &saml.EntityDescriptor{
		EntityID:      s.MetadataURL.String(),
		ValidUntil:    saml.TimeNow().Add(validDuration),
		CacheDuration: validDuration,
		IDPSSODescriptors: []saml.IDPSSODescriptor{
			{
				SSODescriptor: saml.SSODescriptor{
					RoleDescriptor: saml.RoleDescriptor{
						ProtocolSupportEnumeration: "urn:oasis:names:tc:SAML:2.0:protocol",
						KeyDescriptors: []saml.KeyDescriptor{
							{
								Use: "signing",
								KeyInfo: saml.KeyInfo{
									X509Data: saml.X509Data{
										X509Certificates: []saml.X509Certificate{
											{
												Data: certStr,
											},
										},
									},
								},
							},
							{
								Use: "encryption",
								KeyInfo: saml.KeyInfo{
									X509Data: saml.X509Data{
										X509Certificates: []saml.X509Certificate{
											{
												Data: certStr,
											},
										},
									},
								},
								EncryptionMethods: []saml.EncryptionMethod{
									{Algorithm: "http://www.w3.org/2001/04/xmlenc#aes128-cbc"},
									{Algorithm: "http://www.w3.org/2001/04/xmlenc#aes192-cbc"},
									{Algorithm: "http://www.w3.org/2001/04/xmlenc#aes256-cbc"},
									{Algorithm: "http://www.w3.org/2001/04/xmlenc#rsa-oaep-mgf1p"},
								},
							},
						},
					},
					NameIDFormats: []saml.NameIDFormat{saml.NameIDFormat("urn:oasis:names:tc:SAML:2.0:nameid-format:transient")},
				},
				SingleSignOnServices: []saml.Endpoint{
					{
						Binding:  saml.HTTPRedirectBinding,
						Location: s.SSOURL.String(),
					},
					{
						Binding:  saml.HTTPPostBinding,
						Location: s.SSOURL.String(),
					},
				},
			},
		},
	}

	if s.LogoutURL.String() != "" {
		ed.IDPSSODescriptors[0].SSODescriptor.SingleLogoutServices = []saml.Endpoint{
			{
				Binding:  saml.HTTPRedirectBinding,
				Location: s.LogoutURL.String(),
			},
		}
	}

	return ed
}

// Handler returns an http.Handler that serves the metadata and SSO
// URLs
func (s *IdentityProvider) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(s.MetadataURL.Path, s.ServeMetadata)
	mux.HandleFunc(s.SSOURL.Path, s.ServeSSO)
	return mux
}

// ServeMetadata is an http.HandlerFunc that serves the IDP metadata
func (s *IdentityProvider) ServeMetadata(w http.ResponseWriter, r *http.Request) {
	buf, _ := xml.MarshalIndent(s.Metadata(), "", "  ")
	w.Header().Set(headers.ContentType.String(), mediatypes.ApplicationSamlmetadataXml.String())
	w.Write(buf)
}

// ServeSSO handles SAML auth requests.
//
// When it gets a request for a user that does not have a valid session,
// then it prompts the user via XXX.
//
// If the session already exists, then it produces a SAML assertion and
// returns an HTTP response according to the specified binding. The
// only supported binding right now is the HTTP-POST binding which returns
// an HTML form in the appropriate format with Javascript to automatically
// submit that form the to service provider's Assertion Customer Service
// endpoint.
//
// If the SAML request is invalid or cannot be verified a simple StatusBadRequest
// response is sent.
//
// If the assertion cannot be created or returned, a StatusInternalServerError
// response is sent.
func (s *IdentityProvider) ServeSSO(w http.ResponseWriter, r *http.Request) {
	req, err := NewIdpAuthnRequest(s, r)
	if err != nil {
		s.Logger.Printf("failed to parse request: %s", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		s.Logger.Printf("failed to validate request: %s", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// TODO(ross): we must check that the request ID has not been previously
	//   issued.

	session := s.SessionProvider.GetSession(w, r, req)
	if session == nil {
		return
	}

	assertionMaker := s.AssertionMaker
	if assertionMaker == nil {
		assertionMaker = DefaultAssertionMaker{}
	}
	if err := assertionMaker.MakeAssertion(req, session); err != nil {
		s.Logger.Printf("failed to make assertion: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if err := req.WriteResponse(w); err != nil {
		s.Logger.Printf("failed to write response: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

// ServeIDPInitiated handes an IDP-initiated authorization request. Requests of this
// type require us to know a registered service provider and (optionally) the RelayState
// that will be passed to the application.
func (s *IdentityProvider) ServeIDPInitiated(w http.ResponseWriter, r *http.Request, serviceProviderID string, relayState string) {
	req := &IdpAuthnRequest{
		IDP:         s,
		HTTPRequest: r,
		RelayState:  relayState,
		Now:         saml.TimeNow(),
	}

	session := s.SessionProvider.GetSession(w, r, req)
	if session == nil {
		// If GetSession returns nil, it must have written an HTTP response, per the interface
		// (this is probably because it drew a login form or something)
		return
	}

	var err error
	req.ServiceProviderMetadata, err = s.ServiceProviderProvider.GetServiceProvider(r, serviceProviderID)
	if err == os.ErrNotExist {
		s.Logger.Printf("cannot find service provider: %s", serviceProviderID)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	} else if err != nil {
		s.Logger.Printf("cannot find service provider %s: %v", serviceProviderID, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// find an ACS endpoint that we can use
	for _, spssoDescriptor := range req.ServiceProviderMetadata.SPSSODescriptors {
		for _, endpoint := range spssoDescriptor.AssertionConsumerServices {
			if endpoint.Binding == saml.HTTPPostBinding {
				req.ACSEndpoint = &endpoint
				req.SPSSODescriptor = &spssoDescriptor
				break
			}
		}
		if req.ACSEndpoint != nil {
			break
		}
	}
	if req.ACSEndpoint == nil {
		s.Logger.Printf("saml metadata does not contain an Assertion Customer Service url")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	assertionMaker := s.AssertionMaker
	if assertionMaker == nil {
		assertionMaker = DefaultAssertionMaker{}
	}
	if err := assertionMaker.MakeAssertion(req, session); err != nil {
		s.Logger.Printf("failed to make assertion: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err := req.WriteResponse(w); err != nil {
		s.Logger.Printf("failed to write response: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

// IdpAuthnRequest is used by IdentityProvider to handle a single authentication request.
type IdpAuthnRequest struct {
	IDP                     *IdentityProvider
	HTTPRequest             *http.Request
	RelayState              string
	RequestBuffer           []byte
	Request                 saml.AuthnRequest
	ServiceProviderMetadata *saml.EntityDescriptor
	SPSSODescriptor         *saml.SPSSODescriptor
	ACSEndpoint             *saml.IndexedEndpoint
	Assertion               *saml.Assertion
	AssertionEl             *etree.Element
	ResponseEl              *etree.Element
	Now                     time.Time
}

// NewIdpAuthnRequest returns a new IdpAuthnRequest for the given HTTP request to the authorization
// service.
func NewIdpAuthnRequest(idp *IdentityProvider, r *http.Request) (*IdpAuthnRequest, error) {
	req := &IdpAuthnRequest{
		IDP:         idp,
		HTTPRequest: r,
		Now:         saml.TimeNow(),
	}

	switch r.Method {
	case "GET":
		compressedRequest, err := base64.StdEncoding.DecodeString(r.URL.Query().Get("SAMLRequest"))
		if err != nil {
			return nil, fmt.Errorf("cannot decode request: %s", err)
		}
		req.RequestBuffer, err = io.ReadAll(flate.NewReader(bytes.NewReader(compressedRequest)))
		if err != nil {
			return nil, fmt.Errorf("cannot decompress request: %s", err)
		}
		req.RelayState = r.URL.Query().Get("RelayState")
	case "POST":
		if err := r.ParseForm(); err != nil {
			return nil, err
		}
		var err error
		req.RequestBuffer, err = base64.StdEncoding.DecodeString(r.PostForm.Get("SAMLRequest"))
		if err != nil {
			return nil, err
		}
		req.RelayState = r.PostForm.Get("RelayState")
	default:
		return nil, fmt.Errorf("method not allowed")
	}

	return req, nil
}

// Validate checks that the authentication request is valid and assigns
// the AuthnRequest and Metadata properties. Returns a non-nil error if the
// request is not valid.
func (s *IdpAuthnRequest) Validate() error {
	if err := xrv.Validate(bytes.NewReader(s.RequestBuffer)); err != nil {
		return err
	}

	if err := xml.Unmarshal(s.RequestBuffer, &s.Request); err != nil {
		return err
	}

	// We always have exactly one IDP SSO descriptor
	if len(s.IDP.Metadata().IDPSSODescriptors) != 1 {
		panic("expected exactly one IDP SSO descriptor in IDP metadata")
	}
	idpSsoDescriptor := s.IDP.Metadata().IDPSSODescriptors[0]

	// TODO(ross): support signed authn requests
	// For now we do the safe thing and fail in the case where we think
	// requests might be signed.
	if idpSsoDescriptor.WantAuthnRequestsSigned != nil && *idpSsoDescriptor.WantAuthnRequestsSigned {
		return fmt.Errorf("authn request signature checking is not currently supported")
	}

	// In http://docs.oasis-open.org/security/saml/v2.0/saml-bindings-2.0-os.pdf §3.4.5.2
	// we get a description of the Destination attribute:
	//
	//   If the message is signed, the Destination XML attribute in the root SAML
	//   element of the protocol message MUST contain the URL to which the sender
	//   has instructed the user agent to deliver the message. The recipient MUST
	//   then verify that the value matches the location at which the message has
	//   been received.
	//
	// We require the destination be correct either (a) if signing is enabled or
	// (b) if it was provided.
	mustHaveDestination := idpSsoDescriptor.WantAuthnRequestsSigned != nil && *idpSsoDescriptor.WantAuthnRequestsSigned
	mustHaveDestination = mustHaveDestination || s.Request.Destination != ""
	if mustHaveDestination {
		if s.Request.Destination != s.IDP.SSOURL.String() {
			return fmt.Errorf("expected destination to be %q, not %q", s.IDP.SSOURL.String(), s.Request.Destination)
		}
	}

	if s.Request.IssueInstant.Add(saml.MaxIssueDelay).Before(s.Now) {
		return fmt.Errorf("request expired at %s",
			s.Request.IssueInstant.Add(saml.MaxIssueDelay))
	}
	if s.Request.Version != "2.0" {
		return fmt.Errorf("expected SAML request version 2.0 got %v", s.Request.Version)
	}

	// find the service provider
	serviceProviderID := s.Request.Issuer.Value + "/metadata"
	serviceProvider, err := s.IDP.ServiceProviderProvider.GetServiceProvider(s.HTTPRequest, serviceProviderID)
	if err == os.ErrNotExist {
		if s.IDP.AllowUntrustedSP {
			if metadata, err := bhsaml.FetchSAMLMetadata(http.DefaultClient, context.Background(), serviceProviderID); err != nil {
				return fmt.Errorf("cannot fetch metadata from issuer: %s %w", serviceProviderID, err)
			} else {
				s.ServiceProviderMetadata = metadata
			}
		} else {
			return fmt.Errorf("cannot handle request from unknown service provider %s", serviceProviderID)
		}
	} else if err != nil {
		return fmt.Errorf("cannot find service provider %s: %v", serviceProviderID, err)
	} else {
		s.ServiceProviderMetadata = serviceProvider
	}

	// Check that the ACS URL matches an ACS endpoint in the SP metadata.
	if err := s.getACSEndpoint(); err != nil {
		return fmt.Errorf("cannot find assertion consumer service: %v", err)
	}

	return nil
}

func (s *IdpAuthnRequest) getACSEndpoint() error {
	if s.Request.AssertionConsumerServiceIndex != "" {
		for _, spssoDescriptor := range s.ServiceProviderMetadata.SPSSODescriptors {
			for _, spAssertionConsumerService := range spssoDescriptor.AssertionConsumerServices {
				if strconv.Itoa(spAssertionConsumerService.Index) == s.Request.AssertionConsumerServiceIndex {
					s.SPSSODescriptor = &spssoDescriptor
					s.ACSEndpoint = &spAssertionConsumerService
					return nil
				}
			}
		}
	}

	if s.Request.AssertionConsumerServiceURL != "" {
		for _, spssoDescriptor := range s.ServiceProviderMetadata.SPSSODescriptors {
			for _, spAssertionConsumerService := range spssoDescriptor.AssertionConsumerServices {
				if spAssertionConsumerService.Location == s.Request.AssertionConsumerServiceURL {
					s.SPSSODescriptor = &spssoDescriptor
					s.ACSEndpoint = &spAssertionConsumerService
					return nil
				}
			}
		}
	}

	// Some service providers, like the Microsoft Azure AD service provider, issue
	// assertion requests that don't specify an ACS url at all.
	if s.Request.AssertionConsumerServiceURL == "" && s.Request.AssertionConsumerServiceIndex == "" {
		// find a default ACS binding in the metadata that we can use
		for _, spssoDescriptor := range s.ServiceProviderMetadata.SPSSODescriptors {
			for _, spAssertionConsumerService := range spssoDescriptor.AssertionConsumerServices {
				if spAssertionConsumerService.IsDefault != nil && *spAssertionConsumerService.IsDefault {
					switch spAssertionConsumerService.Binding {
					case saml.HTTPPostBinding, saml.HTTPRedirectBinding:
						s.SPSSODescriptor = &spssoDescriptor
						s.ACSEndpoint = &spAssertionConsumerService
						return nil
					}
				}
			}
		}

		// if we can't find a default, use *any* ACS binding
		for _, spssoDescriptor := range s.ServiceProviderMetadata.SPSSODescriptors {
			for _, spAssertionConsumerService := range spssoDescriptor.AssertionConsumerServices {
				switch spAssertionConsumerService.Binding {
				case saml.HTTPPostBinding, saml.HTTPRedirectBinding:
					s.SPSSODescriptor = &spssoDescriptor
					s.ACSEndpoint = &spAssertionConsumerService
					return nil
				}
			}
		}
	}

	return os.ErrNotExist // no ACS url found or specified
}

// DefaultAssertionMaker produces a SAML assertion for the
// given request and assigns it to req.Assertion.
type DefaultAssertionMaker struct {
}

// MakeAssertion implements AssertionMaker. It produces a SAML assertion from the
// given request and assigns it to req.Assertion.
func (DefaultAssertionMaker) MakeAssertion(req *IdpAuthnRequest, session *Session) error {
	attributes := []saml.Attribute{}

	var attributeConsumingService *saml.AttributeConsumingService
	for _, acs := range req.SPSSODescriptor.AttributeConsumingServices {
		if acs.IsDefault != nil && *acs.IsDefault {
			attributeConsumingService = &acs
			break
		}
	}
	if attributeConsumingService == nil {
		for _, acs := range req.SPSSODescriptor.AttributeConsumingServices {
			attributeConsumingService = &acs
			break
		}
	}
	if attributeConsumingService == nil {
		attributeConsumingService = &saml.AttributeConsumingService{}
	}

	for _, requestedAttribute := range attributeConsumingService.RequestedAttributes {
		if requestedAttribute.NameFormat == "urn:oasis:names:tc:SAML:2.0:attrname-format:basic" || requestedAttribute.NameFormat == "urn:oasis:names:tc:SAML:2.0:attrname-format:unspecified" {
			attrName := requestedAttribute.Name
			attrName = regexp.MustCompile("[^A-Za-z0-9]+").ReplaceAllString(attrName, "")
			switch attrName {
			case "email", "emailaddress":
				attributes = append(attributes, saml.Attribute{
					FriendlyName: requestedAttribute.FriendlyName,
					Name:         requestedAttribute.Name,
					NameFormat:   requestedAttribute.NameFormat,
					Values: []saml.AttributeValue{{
						Type:  "xs:string",
						Value: session.UserEmail,
					}},
				})
			case "name", "fullname", "cn", "commonname":
				attributes = append(attributes, saml.Attribute{
					FriendlyName: requestedAttribute.FriendlyName,
					Name:         requestedAttribute.Name,
					NameFormat:   requestedAttribute.NameFormat,
					Values: []saml.AttributeValue{{
						Type:  "xs:string",
						Value: session.UserCommonName,
					}},
				})
			case "givenname", "firstname":
				attributes = append(attributes, saml.Attribute{
					FriendlyName: requestedAttribute.FriendlyName,
					Name:         requestedAttribute.Name,
					NameFormat:   requestedAttribute.NameFormat,
					Values: []saml.AttributeValue{{
						Type:  "xs:string",
						Value: session.UserGivenName,
					}},
				})
			case "surname", "lastname", "familyname":
				attributes = append(attributes, saml.Attribute{
					FriendlyName: requestedAttribute.FriendlyName,
					Name:         requestedAttribute.Name,
					NameFormat:   requestedAttribute.NameFormat,
					Values: []saml.AttributeValue{{
						Type:  "xs:string",
						Value: session.UserSurname,
					}},
				})
			case "uid", "user", "userid":
				attributes = append(attributes, saml.Attribute{
					FriendlyName: requestedAttribute.FriendlyName,
					Name:         requestedAttribute.Name,
					NameFormat:   requestedAttribute.NameFormat,
					Values: []saml.AttributeValue{{
						Type:  "xs:string",
						Value: session.UserName,
					}},
				})
			}
		}
	}

	if session.UserName != "" {
		attributes = append(attributes, saml.Attribute{
			FriendlyName: "uid",
			Name:         "urn:oid:0.9.2342.19200300.100.1.1",
			NameFormat:   "urn:oasis:names:tc:SAML:2.0:attrname-format:uri",
			Values: []saml.AttributeValue{{
				Type:  "xs:string",
				Value: session.UserName,
			}},
		})
	}

	if session.UserEmail != "" {
		attributes = append(attributes, saml.Attribute{
			FriendlyName: "eduPersonPrincipalName",
			Name:         bhsaml.ObjectIDEmail,
			NameFormat:   bhsaml.ObjectIDAttributeNameFormat,
			Values: []saml.AttributeValue{{
				Type:  "xs:string",
				Value: session.UserEmail,
			}},
		})
	}
	if session.UserSurname != "" {
		attributes = append(attributes, saml.Attribute{
			FriendlyName: "sn",
			Name:         "urn:oid:2.5.4.4",
			NameFormat:   "urn:oasis:names:tc:SAML:2.0:attrname-format:uri",
			Values: []saml.AttributeValue{{
				Type:  "xs:string",
				Value: session.UserSurname,
			}},
		})
	}
	if session.UserGivenName != "" {
		attributes = append(attributes, saml.Attribute{
			FriendlyName: "givenName",
			Name:         "urn:oid:2.5.4.42",
			NameFormat:   "urn:oasis:names:tc:SAML:2.0:attrname-format:uri",
			Values: []saml.AttributeValue{{
				Type:  "xs:string",
				Value: session.UserGivenName,
			}},
		})
	}

	if session.UserCommonName != "" {
		attributes = append(attributes, saml.Attribute{
			FriendlyName: "cn",
			Name:         "urn:oid:2.5.4.3",
			NameFormat:   "urn:oasis:names:tc:SAML:2.0:attrname-format:uri",
			Values: []saml.AttributeValue{{
				Type:  "xs:string",
				Value: session.UserCommonName,
			}},
		})
	}

	if session.UserScopedAffiliation != "" {
		attributes = append(attributes, saml.Attribute{
			FriendlyName: "uid",
			Name:         "urn:oid:1.3.6.1.4.1.5923.1.1.1.9",
			NameFormat:   "urn:oasis:names:tc:SAML:2.0:attrname-format:uri",
			Values: []saml.AttributeValue{{
				Type:  "xs:string",
				Value: session.UserScopedAffiliation,
			}},
		})
	}

	attributes = append(attributes, session.CustomAttributes...)

	if len(session.Groups) != 0 {
		groupMemberAttributeValues := []saml.AttributeValue{}
		for _, group := range session.Groups {
			groupMemberAttributeValues = append(groupMemberAttributeValues, saml.AttributeValue{
				Type:  "xs:string",
				Value: group,
			})
		}
		attributes = append(attributes, saml.Attribute{
			FriendlyName: "eduPersonAffiliation",
			Name:         "urn:oid:1.3.6.1.4.1.5923.1.1.1.1",
			NameFormat:   "urn:oasis:names:tc:SAML:2.0:attrname-format:uri",
			Values:       groupMemberAttributeValues,
		})
	}

	// allow for some clock skew in the validity period using the
	// issuer's apparent clock.
	notBefore := req.Now.Add(-1 * saml.MaxClockSkew)
	notOnOrAfterAfter := req.Now.Add(saml.MaxIssueDelay)
	if notBefore.Before(req.Request.IssueInstant) {
		notBefore = req.Request.IssueInstant
		notOnOrAfterAfter = notBefore.Add(saml.MaxIssueDelay)
	}

	req.Assertion = &saml.Assertion{
		ID:           fmt.Sprintf("id-%x", randomBytes(20)),
		IssueInstant: saml.TimeNow(),
		Version:      "2.0",
		Issuer: saml.Issuer{
			Format: "urn:oasis:names:tc:SAML:2.0:nameid-format:entity",
			Value:  req.IDP.Metadata().EntityID,
		},
		Subject: &saml.Subject{
			NameID: &saml.NameID{
				Format:          "urn:oasis:names:tc:SAML:2.0:nameid-format:transient",
				NameQualifier:   req.IDP.Metadata().EntityID,
				SPNameQualifier: req.ServiceProviderMetadata.EntityID,
				Value:           session.NameID,
			},
			SubjectConfirmations: []saml.SubjectConfirmation{
				{
					Method: "urn:oasis:names:tc:SAML:2.0:cm:bearer",
					SubjectConfirmationData: &saml.SubjectConfirmationData{
						Address:      req.HTTPRequest.RemoteAddr,
						InResponseTo: req.Request.ID,
						NotOnOrAfter: req.Now.Add(saml.MaxIssueDelay),
						Recipient:    req.ACSEndpoint.Location,
					},
				},
			},
		},
		Conditions: &saml.Conditions{
			NotBefore:    notBefore,
			NotOnOrAfter: notOnOrAfterAfter,
			AudienceRestrictions: []saml.AudienceRestriction{
				{
					Audience: saml.Audience{Value: req.ServiceProviderMetadata.EntityID},
				},
			},
		},
		AuthnStatements: []saml.AuthnStatement{
			{
				AuthnInstant: session.CreateTime,
				SessionIndex: session.Index,
				SubjectLocality: &saml.SubjectLocality{
					Address: req.HTTPRequest.RemoteAddr,
				},
				AuthnContext: saml.AuthnContext{
					AuthnContextClassRef: &saml.AuthnContextClassRef{
						Value: "urn:oasis:names:tc:SAML:2.0:ac:classes:PasswordProtectedTransport",
					},
				},
			},
		},
		AttributeStatements: []saml.AttributeStatement{
			{
				Attributes: attributes,
			},
		},
	}

	return nil
}

// The Canonicalizer prefix list MUST be empty. Various implementations
// (maybe ours?) do not appear to support non-empty prefix lists in XML C14N.
const canonicalizerPrefixList = ""

// MakeAssertionEl sets `AssertionEl` to a signed, possibly encrypted, version of `Assertion`.
func (s *IdpAuthnRequest) MakeAssertionEl() error {
	keyPair := tls.Certificate{
		Certificate: [][]byte{s.IDP.Certificate.Raw},
		PrivateKey:  s.IDP.Key,
		Leaf:        s.IDP.Certificate,
	}
	for _, cert := range s.IDP.Intermediates {
		keyPair.Certificate = append(keyPair.Certificate, cert.Raw)
	}
	keyStore := dsig.TLSCertKeyStore(keyPair)

	signatureMethod := s.IDP.SignatureMethod
	if signatureMethod == "" {
		signatureMethod = dsig.RSASHA1SignatureMethod
	}

	signingContext := dsig.NewDefaultSigningContext(keyStore)
	signingContext.Canonicalizer = dsig.MakeC14N10ExclusiveCanonicalizerWithPrefixList(canonicalizerPrefixList)
	if err := signingContext.SetSignatureMethod(signatureMethod); err != nil {
		return err
	}

	assertionEl := s.Assertion.Element()

	signedAssertionEl, err := signingContext.SignEnveloped(assertionEl)
	if err != nil {
		return err
	}

	sigEl := signedAssertionEl.Child[len(signedAssertionEl.Child)-1]
	s.Assertion.Signature = sigEl.(*etree.Element)
	signedAssertionEl = s.Assertion.Element()

	certBuf, err := s.getSPEncryptionCert()
	if err == os.ErrNotExist {
		s.AssertionEl = signedAssertionEl
		return nil
	} else if err != nil {
		return err
	}

	var signedAssertionBuf []byte
	{
		doc := etree.NewDocument()
		doc.SetRoot(signedAssertionEl)
		signedAssertionBuf, err = doc.WriteToBytes()
		if err != nil {
			return err
		}
	}

	encryptor := xmlenc.OAEP()
	encryptor.BlockCipher = xmlenc.AES128CBC
	encryptor.DigestMethod = &xmlenc.SHA1
	encryptedDataEl, err := encryptor.Encrypt(certBuf, signedAssertionBuf, nil)
	if err != nil {
		return err
	}
	encryptedDataEl.CreateAttr("Type", "http://www.w3.org/2001/04/xmlenc#Element")

	encryptedAssertionEl := etree.NewElement("saml:EncryptedAssertion")
	encryptedAssertionEl.AddChild(encryptedDataEl)
	s.AssertionEl = encryptedAssertionEl

	return nil
}

// WriteResponse writes the `Response` to the http.ResponseWriter. If
// `Response` is not already set, it calls MakeResponse to produce it.
func (s *IdpAuthnRequest) WriteResponse(w http.ResponseWriter) error {
	if s.ResponseEl == nil {
		if err := s.MakeResponse(); err != nil {
			return err
		}
	}

	doc := etree.NewDocument()
	doc.SetRoot(s.ResponseEl)
	responseBuf, err := doc.WriteToBytes()
	if err != nil {
		return err
	}

	// the only supported binding is the HTTP-POST binding
	switch s.ACSEndpoint.Binding {
	case saml.HTTPPostBinding:
		tmpl := template.Must(template.New("saml-post-form").Parse(`<html>` +
			`<form method="post" action="{{.URL}}" id="SAMLResponseForm">` +
			`<input type="hidden" name="SAMLResponse" value="{{.SAMLResponse}}" />` +
			`<input type="hidden" name="RelayState" value="{{.RelayState}}" />` +
			`<input id="SAMLSubmitButton" type="submit" value="Continue" />` +
			`</form>` +
			`<script>document.getElementById('SAMLSubmitButton').style.visibility='hidden';</script>` +
			`<script>document.getElementById('SAMLResponseForm').submit();</script>` +
			`</html>`))
		data := struct {
			URL          string
			SAMLResponse string
			RelayState   string
		}{
			URL:          s.ACSEndpoint.Location,
			SAMLResponse: base64.StdEncoding.EncodeToString(responseBuf),
			RelayState:   s.RelayState,
		}

		buf := bytes.NewBuffer(nil)
		if err := tmpl.Execute(buf, data); err != nil {
			return err
		}
		if _, err := io.Copy(w, buf); err != nil {
			return err
		}
		return nil

	default:
		return fmt.Errorf("%s: unsupported binding %s",
			s.ServiceProviderMetadata.EntityID,
			s.ACSEndpoint.Binding)
	}
}

// getSPEncryptionCert returns the certificate which we can use to encrypt things
// to the SP in PEM format, or nil if no such certificate is found.
func (s *IdpAuthnRequest) getSPEncryptionCert() (*x509.Certificate, error) {
	certStr := ""
	for _, keyDescriptor := range s.SPSSODescriptor.KeyDescriptors {
		if keyDescriptor.Use == "encryption" {
			certStr = keyDescriptor.KeyInfo.X509Data.X509Certificates[0].Data
			break
		}
	}

	// If there are no certs explicitly labeled for encryption, return the first
	// non-empty cert we find.
	if certStr == "" {
		for _, keyDescriptor := range s.SPSSODescriptor.KeyDescriptors {
			if keyDescriptor.Use == "" && len(keyDescriptor.KeyInfo.X509Data.X509Certificates) > 0 {
				certStr = keyDescriptor.KeyInfo.X509Data.X509Certificates[0].Data
				break
			}
		}
	}

	if certStr == "" {
		return nil, os.ErrNotExist
	}

	// cleanup whitespace and re-encode a PEM
	certStr = regexp.MustCompile(`\s+`).ReplaceAllString(certStr, "")
	certBytes, err := base64.StdEncoding.DecodeString(certStr)
	if err != nil {
		return nil, fmt.Errorf("cannot decode certificate base64: %v", err)
	}
	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, fmt.Errorf("cannot parse certificate: %v", err)
	}
	return cert, nil
}

// MakeResponse creates and assigns a new SAML response in ResponseEl. `Assertion` must
// be non-nil. If MakeAssertionEl() has not been called, this function calls it for
// you.
func (s *IdpAuthnRequest) MakeResponse() error {
	if s.AssertionEl == nil {
		if err := s.MakeAssertionEl(); err != nil {
			return err
		}
	}

	response := &saml.Response{
		Destination:  s.ACSEndpoint.Location,
		ID:           fmt.Sprintf("id-%x", randomBytes(20)),
		InResponseTo: s.Request.ID,
		IssueInstant: s.Now,
		Version:      "2.0",
		Issuer: &saml.Issuer{
			Format: "urn:oasis:names:tc:SAML:2.0:nameid-format:entity",
			Value:  s.IDP.MetadataURL.String(),
		},
		Status: saml.Status{
			StatusCode: saml.StatusCode{
				Value: saml.StatusSuccess,
			},
		},
	}

	responseEl := response.Element()
	responseEl.AddChild(s.AssertionEl) // AssertionEl either an EncryptedAssertion or Assertion element

	// Sign the response element (we've already signed the Assertion element)
	{
		keyPair := tls.Certificate{
			Certificate: [][]byte{s.IDP.Certificate.Raw},
			PrivateKey:  s.IDP.Key,
			Leaf:        s.IDP.Certificate,
		}
		for _, cert := range s.IDP.Intermediates {
			keyPair.Certificate = append(keyPair.Certificate, cert.Raw)
		}
		keyStore := dsig.TLSCertKeyStore(keyPair)

		signatureMethod := s.IDP.SignatureMethod
		if signatureMethod == "" {
			signatureMethod = dsig.RSASHA1SignatureMethod
		}

		signingContext := dsig.NewDefaultSigningContext(keyStore)
		signingContext.Canonicalizer = dsig.MakeC14N10ExclusiveCanonicalizerWithPrefixList(canonicalizerPrefixList)
		if err := signingContext.SetSignatureMethod(signatureMethod); err != nil {
			return err
		}

		signedResponseEl, err := signingContext.SignEnveloped(responseEl)
		if err != nil {
			return err
		}

		sigEl := signedResponseEl.ChildElements()[len(signedResponseEl.ChildElements())-1]
		response.Signature = sigEl
		responseEl = response.Element()
		responseEl.AddChild(s.AssertionEl)
	}

	s.ResponseEl = responseEl
	return nil
}
