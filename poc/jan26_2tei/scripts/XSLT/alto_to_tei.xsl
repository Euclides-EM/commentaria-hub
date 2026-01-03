<?xml version="1.0" encoding="UTF-8"?>
<xsl:stylesheet xmlns:xsl="http://www.w3.org/1999/XSL/Transform"
    xmlns:xs="http://www.w3.org/2001/XMLSchema"
    xmlns:alto="http://www.loc.gov/standards/alto/ns-v4#"
    xmlns:tei="http://www.tei-c.org/ns/1.0"
    exclude-result-prefixes="xs alto"
    version="1.0">

    <xsl:output method="xml" indent="yes" encoding="UTF-8"/>

    <xsl:template match="/">
        <TEI xmlns="http://www.tei-c.org/ns/1.0">
            <teiHeader>
                <fileDesc>
                    <titleStmt>
                        <title>Converted from ALTO</title>
                    </titleStmt>
                    <publicationStmt>
                        <p>Publication Information</p>
                    </publicationStmt>
                    <sourceDesc>
                        <p>Information about the source</p>
                    </sourceDesc>
                </fileDesc>
            </teiHeader>
            <text>
                <body>
                    <xsl:apply-templates select="//alto:Page"/>
                </body>
            </text>
        </TEI>
    </xsl:template>

    <xsl:template match="alto:Page">
        <pb facs="{@ID}"/>
        <xsl:apply-templates select=".//alto:TextBlock"/>
    </xsl:template>

    <xsl:template match="alto:TextBlock">
        <p>
            <xsl:apply-templates select=".//alto:TextLine"/>
        </p>
    </xsl:template>

    <xsl:template match="alto:TextLine">
        <xsl:apply-templates select="alto:String"/>
        <xsl:if test="following-sibling::alto:TextLine">
            <lb/>
        </xsl:if>
    </xsl:template>

    <xsl:template match="alto:String">
        <xsl:value-of select="@CONTENT"/>
        <xsl:text> </xsl:text>
    </xsl:template>

</xsl:stylesheet>